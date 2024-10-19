#!/bin/bash

# Get the first argument passed to the script and store it as project name
if [ $# -eq 0 ]; then
    echo "Error: No project name provided. Please provide a project name as an argument."
    exit 1
fi

PROJECT_NAME="$1"

# Get the second argument passed to the script and store it as region
if [ $# -lt 2 ]; then
    echo "Error: No region provided. Please provide a project name and region as arguments."
    exit 1
fi

REGION="$2"

# Set the project in gcloud
gcloud config set project "$PROJECT_NAME"

# Verify the project is set correctly
echo "Deploying to project: $PROJECT_NAME"
# Verify the region is set correctly
echo "Deploying to region: $REGION"


# Build the Docker image using Cloud Build
gcloud builds submit --tag gcr.io/$PROJECT_NAME/courses-api .

gcloud run deploy courses-api --allow-unauthenticated --region $REGION --quiet --image gcr.io/$PROJECT_NAME/courses-api
 