
docker build -t pycon-course-api:latest .

docker run -p 8080:8080  -d -it --rm --name pycon-api pycon-course-api:latest 

docker stop pycon-api

# Enable Artifact Registry API in your GCP project
# Replace [PROJECT_ID] with your actual GCP project ID

gcloud services enable artifactregistry.googleapis.com --project=[PROJECT_ID]


