curl -X POST      -H "Authorization: Bearer $(gcloud auth print-access-token)"      -H "Content-Type: application/json; charset=utf-8"      -d '{
        "input": {
            "input": "Can you please share what are being taught on this course", 
            "session_id": "fe66e55f-f107-493b-980a-845a1000d9ee"
        }
    }' "https://us-central1-aiplatform.googleapis.com/v1beta1/projects/imrenagi-gemini-experiment/locations/us-central1/reasoningEngines/5811983280051847168:query"