#!/bin/bash

gcloud run deploy course-agent-ui --env-vars-file=vars.yml  --allow-unauthenticated --source . --region us-central1
