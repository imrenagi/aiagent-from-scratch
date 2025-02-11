package main

import (
	"context"
	"encoding/json"
	"html/template"
	"time"

	"net/http"
	"voice-agent/courses"
	"voice-agent/interviews"

	_ "embed"

	gogenai "github.com/google/generative-ai-go/genai"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/genai"
)

var upgrader = websocket.Upgrader{}

const (
	modelName          = "gemini-2.0-flash-exp"
	embeddingModelName = "text-embedding-004"
)

//go:embed index.html
var homeTemplate string

func (s Server) voiceChaHandler(model string, cfg *genai.LiveConnectConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error().Err(err).Msg("upgrade websocket error")
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		defer c.Close()

		session, err := s.GenAIClient.Live.Connect(model, cfg)
		if err != nil {
			log.Error().Err(err).Msg("unable to start live session")
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		defer session.Close()

		errChan := make(chan error)
		doneChan := make(chan struct{})

		// Get model's response
		go func() {
			for {
				message, err := session.Receive()
				if err != nil {
					log.Error().Err(err).Msg("receive error on live session response")
					errChan <- err
					return
				}

				var functionResponses []*genai.FunctionResponse
				if message.ToolCall != nil {
					functionCalls := message.ToolCall.FunctionCalls
					for _, fc := range functionCalls {
						log.Debug().Str("name", fc.Name).
							Str("id", fc.ID).
							Any("params", fc.Args).
							Msg("checking function call")
						fr, err := s.Dispatch(r.Context(), fc)
						if err != nil {
							log.Error().Err(err).Msg("dispatch error")
							errChan <- err
							return
						}
						functionResponses = append(functionResponses, fr)
					}
					log.Debug().Msg("sending tool response")
					err := session.Send(&genai.LiveClientMessage{
						ToolResponse: &genai.LiveClientToolResponse{
							FunctionResponses: functionResponses,
						},
					})
					if err != nil {
						log.Error().Err(err).Msg("send tool response error")
						errChan <- err
						return
					}
					log.Debug().Msg("tool response sent")
				}

				messageBytes, err := json.Marshal(message)
				if err != nil {
					log.Error().Err(err).Msg("marshal model response error")
					errChan <- err
					return
				}
				err = c.WriteMessage(websocket.TextMessage, messageBytes)
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
						log.Error().Err(err).Msg("got unexpected websocket close error in write")
						errChan <- err
					} else {
						log.Debug().Err(err).Msg("websocket closed in write")
						doneChan <- struct{}{}
					}
					return
				}
			}
		}()

		go func() {
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
						log.Error().Err(err).Msg("got unexpected websocket close error in read")
						errChan <- err
					} else {
						log.Debug().Err(err).Msg("websocket closed in read")
						doneChan <- struct{}{}
					}
					return
				}

				// TODO currently the input is genai.LiveClientMessage,
				// we can create different contract with our client
				// and convert it to genai.LiveClientMessage later
				var sendMessage genai.LiveClientMessage
				if err := json.Unmarshal(message, &sendMessage); err != nil {
					log.Error().Err(err).Msg("unmarshal message error")
					errChan <- err
					return
				}

				if err := session.Send(&sendMessage); err != nil {
					log.Error().Err(err).Msg("send message to session error")
					errChan <- err
					return
				}
			}
		}()

		for {
			select {
			case <-doneChan:
				log.Debug().Msg("function done")
				return
			case err, ok := <-errChan:
				if !ok {
					log.Warn().Msg("error channel closed")
					return
				} else {
					writeError(w, http.StatusInternalServerError, err)
					return
				}
			}
		}
	}
}

func (s Server) CourseVoiceChaHandler() http.HandlerFunc {
	config := &genai.LiveConnectConfig{
		GenerationConfig: &genai.GenerationConfig{
			AudioTimestamp: true,
		},
		ResponseModalities: []genai.Modality{
			genai.ModalityAudio,
			genai.ModalityText,
		},
		SpeechConfig: &genai.SpeechConfig{
			VoiceConfig: &genai.VoiceConfig{
				PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
					VoiceName: "Kore",
				},
			},
		},
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{
					Text: courses.SystemPrompt,
				},
			},
		},
		Tools: courses.Tools,
	}
	return s.voiceChaHandler(modelName, config)
}

func (s Server) InterviewerVoiceChaHandler() http.HandlerFunc {
	config := &genai.LiveConnectConfig{
		GenerationConfig: &genai.GenerationConfig{
			AudioTimestamp: true,
		},
		ResponseModalities: []genai.Modality{
			genai.ModalityAudio,
			genai.ModalityText,
		},
		SpeechConfig: &genai.SpeechConfig{
			VoiceConfig: &genai.VoiceConfig{
				PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
					VoiceName: "Kore",
				},
			},
		},
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{
					Text: interviews.SystemPrompt,
				},
			},
		},
	}
	return s.voiceChaHandler(modelName, config)
}

func (s Server) CourseAgent() http.HandlerFunc {
	return s.voiceChatPage("/api/v1/courses/voice_sessions:start")
}

func (s Server) InterviewAgent() http.HandlerFunc {
	return s.voiceChatPage("/api/v1/interviewers/voice_sessions:start")
}

func (s Server) voiceChatPage(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("home").Parse(homeTemplate)
		if err != nil {
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, "ws://"+r.Host+path)
		if err != nil {
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
	}
}

type Server struct {
	GenAIClient    *genai.Client
	EmbeddingModel *gogenai.EmbeddingModel
	DB             *sqlx.DB
	VectorStore    *VectorStore
}

func (s *Server) Start(ctx context.Context) {
	mux := mux.NewRouter()
	mux.Handle("/courses", s.CourseAgent())
	mux.Handle("/interviews", s.InterviewAgent())

	api := mux.PathPrefix("/api").Subrouter()

	coursesV1Router := api.PathPrefix("/v1/courses").Subrouter()
	coursesV1Router.Handle("/voice_sessions:start", s.CourseVoiceChaHandler())

	interviewersV1Router := api.PathPrefix("/v1/interviewers").Subrouter()
	interviewersV1Router.Handle("/voice_sessions:start", s.InterviewerVoiceChaHandler())

	server := &http.Server{
		Addr:              ":8000",
		Handler:           mux,
		ReadTimeout:       15 * time.Minute,
		WriteTimeout:      15 * time.Minute,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       10 * time.Second,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal().Msgf("failed to start server: %v", err)
		}
	}()

	<-ctx.Done()

	gracefulShutdownPeriod := 5 * time.Second

	log.Warn().Msg("shutting down http server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownPeriod)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("failed to shutdown http server gracefully")
	}
}

type cError struct {
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	b, _ := json.Marshal(cError{Message: err.Error()})
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
