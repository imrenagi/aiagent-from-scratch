package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"

	gogenai "github.com/google/generative-ai-go/genai"
	"google.golang.org/genai"
)

var systemPrompt = `You are a bot assistant that sells online course about software security. You only use information provided from datastore or tools. You can provide the information that is relevant to the user's question or the summary of the content. If they ask about the content, you can give them more detail about the content. If the user seems interested, you may suggest the user to enroll in the course.`

var courseAgentTools = []*genai.Tool{
	{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        "search_course_content",
				Description: "Explain about software security course materials.",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"query": {
							Type:        genai.TypeString,
							Description: "search query to search course content.",
						},
					},
					Required: []string{"query"},
				},
			},
			{
				Name:        "list_courses",
				Description: "List all available courses sold on the platform.",
			},
			{
				Name:        "get_course",
				Description: "Get course details by course name. course name is the unique identifier of the course. it typically contains the course title with dashes. This function can be used to get course details such as course price, etc.",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"course": {
							Type:        genai.TypeString,
							Description: "name of the course. this is the unique identifier of the course. it typically contains the course title with dashes, all in lowercase.",
						},
					},
					Required: []string{"course"},
				},
			},
			{
				Name:        "create_order",
				Description: "Create order for a course. This function can be used to create an order for a course. When this function returns successfully, it will return payment url to user to make payment.",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"course": {
							Type:        genai.TypeString,
							Description: "name of the course. this is the unique identifier of the course. it typically contains the course title with dashes, all in lowercase.",
						},
						"user_name": {
							Type:        genai.TypeString,
							Description: "name of the user who is purchasing the course .",
						},
						"user_email": {
							Type:        genai.TypeString,
							Description: "email of the user who is purchasing the course .",
						},
					},
					Required: []string{"course", "user_name", "user_email"},
				},
			},
			{
				Name:        "get_order",
				Description: "Get order by using order number. This function can be used to get order details such as payment status to check whether the order has been paid or not. If user already paid the course, say thanks",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"order_number": {
							Type:        genai.TypeString,
							Description: "order number identifier. this is a unique identifier in uuid format.",
						},
					},
					Required: []string{"order_number"},
				},
			},
		},
	},
}

type ApiTool struct {
}

type VectorTool struct {
}

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
	o, err := GetOrder(ctx, orderNumber)
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
	o, err := CreateOrder(ctx, course, userName, userEmail)
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

	c, err := GetCourse(ctx, course)
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
	courses, err := ListCourse(ctx)
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
