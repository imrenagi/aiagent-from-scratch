package courses

import (
	"google.golang.org/genai"
)

var SystemPrompt = `You are a bot assistant that sells online course about software security. You only use information provided from datastore or tools. You can provide the information that is relevant to the user's question or the summary of the content. If they ask about the content, you can give them more detail about the content. If the user seems interested, you may suggest the user to enroll in the course.`

var Tools = []*genai.Tool{
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
