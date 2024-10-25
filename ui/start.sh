#!/bin/bash

pip install -r requirements.txt

export UI_PROJECT_ID="pyconapac2024-testing"
export UI_LOCATION="us-central1"  
export UI_STAGING_BUCKET="gs://pyconapac24-staging-bucket-01"  
export UI_REASONING_ENGINE_PATH="projects/423999378850/locations/us-central1/reasoningEngines/797013988742266880"

streamlit run server.py --server.port=8080 --server.address=0.0.0.0