package main

import (
	"context"
	"encoding/json"
	"fmt"

	"voice-agent/courses"

	"github.com/rs/zerolog/log"

	gogenai "github.com/google/generative-ai-go/genai"
	"google.golang.org/genai"
)

func (s Server) Dispatch(ctx context.Context, fc *genai.FunctionCall) (*genai.FunctionResponse, error) {
	switch fc.Name {
	case "list_courses":
		return s.ListCourses(ctx)
	case "get_course":
		return s.GetCourse(ctx, fc)
	case "create_order":
		return s.CreateOrder(ctx, fc)
	case "get_order":
		return s.GetOrder(ctx, fc)
	case "search_course_content":
		return s.SearchCourseContent(ctx, fc.Args["query"].(string))
	default:
		return nil, fmt.Errorf("unknown function %s", fc.Name)
	}
}

func (s Server) GetOrder(ctx context.Context, fc *genai.FunctionCall) (*genai.FunctionResponse, error) {
	orderNumber, ok := fc.Args["order_number"].(string)
	if !ok {
		return nil, fmt.Errorf("missing order number")
	}
	o, err := courses.GetOrder(ctx, orderNumber)
	if err != nil {
		return nil, err
	}
	var om map[string]any
	b, _ := json.Marshal(o)
	err = json.Unmarshal(b, &om)
	if err != nil {
		return nil, err
	}
	log.Debug().Interface("order", om).Msg("checking order")
	fr := &genai.FunctionResponse{
		Name: "get_order",
		Response: map[string]interface{}{
			"order": om,
		},
	}
	return fr, nil
}

func (s Server) CreateOrder(ctx context.Context, fc *genai.FunctionCall) (*genai.FunctionResponse, error) {
	course, ok := fc.Args["course"].(string)
	if !ok {
		return nil, fmt.Errorf("missing course")
	}
	userName, ok := fc.Args["user_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing user name")
	}
	userEmail, ok := fc.Args["user_email"].(string)
	if !ok {
		return nil, fmt.Errorf("missing user email")
	}
	o, err := courses.CreateOrder(ctx, course, userName, userEmail)
	if err != nil {
		return nil, err
	}

	var om map[string]any
	b, _ := json.Marshal(o)
	err = json.Unmarshal(b, &om)
	if err != nil {
		return nil, err
	}
	log.Debug().Interface("order", om).Msg("checking order")

	paymentUrl := fmt.Sprintf("http://localhost:8080/orders/%s/payment", o.ID)
	log.Info().Str("payment_url", paymentUrl).Msg("payment url")

	return &genai.FunctionResponse{
		Name: "create_order",
		Response: map[string]interface{}{
			"order":       om,
			"payment_url": paymentUrl,
		},
	}, nil
}

func (s Server) GetCourse(ctx context.Context, fc *genai.FunctionCall) (*genai.FunctionResponse, error) {
	course, ok := fc.Args["course"].(string)
	if !ok {
		return nil, fmt.Errorf("missing course")
	}

	c, err := courses.GetCourse(ctx, course)
	if err != nil {
		return nil, err
	}
	var cm map[string]any
	b, _ := json.Marshal(c)
	err = json.Unmarshal(b, &cm)
	if err != nil {
		return nil, err
	}
	log.Debug().Interface("course", cm).Msg("checking course")
	fr := &genai.FunctionResponse{
		Name: "get_course",
		Response: map[string]interface{}{
			"course": cm,
		},
	}
	return fr, nil
}

func (s Server) ListCourses(ctx context.Context) (*genai.FunctionResponse, error) {
	courses, err := courses.ListCourse(ctx)
	if err != nil {
		return nil, err
	}
	var coursesM []map[string]interface{}
	b, _ := json.Marshal(courses)
	err = json.Unmarshal(b, &coursesM)
	if err != nil {
		return nil, err
	}
	log.Debug().Interface("courses", coursesM).Msg("checking courses")
	fr := &genai.FunctionResponse{
		Name: "list_courses",
		Response: map[string]interface{}{
			"courses": coursesM,
		},
	}
	return fr, nil
}

func (s Server) SearchCourseContent(ctx context.Context, query string) (*genai.FunctionResponse, error) {
	resp, err := s.EmbeddingModel.EmbedContent(ctx, gogenai.Text(query))
	if err != nil {
		return nil, err
	}
	data, err := s.VectorStore.QueryContent(ctx, resp.Embedding.Values, 0)
	if err != nil {
		return nil, err
	}
	fr := &genai.FunctionResponse{
		Name: "search_course_content",
		Response: map[string]interface{}{
			"contents": data,
		},
	}
	return fr, nil
}
