#!/bin/bash

# pip install -r requirements.txt

export UI_PROJECT_ID="imrenagi-devfest-2024"
export UI_LOCATION="us-central1"  
export UI_STAGING_BUCKET="gs://devfest24-demo-bucket"  
export UI_REASONING_ENGINE_PATH="projects/908311267620/locations/us-central1/reasoningEngines/2387523529017917440"

streamlit run server.py --server.port=8080 --server.address=0.0.0.0