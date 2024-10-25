import streamlit as st
import uuid
import vertexai
import os
from vertexai.preview import reasoning_engines

PROJECT_ID = os.getenv("UI_PROJECT_ID") 
LOCATION = os.getenv("UI_LOCATION")
STAGING_BUCKET = os.getenv("UI_STAGING_BUCKET")
REASONING_ENGINE_PATH = os.getenv("UI_REASONING_ENGINE_PATH")

vertexai.init(project=PROJECT_ID, location=LOCATION, staging_bucket=STAGING_BUCKET)
remote_agent = reasoning_engines.ReasoningEngine(REASONING_ENGINE_PATH)

st.title("Course Agent")

def get_response(input: str):    
    response = remote_agent.query(
    input=input,
    session_id=st.session_state.session_id,
    )
    return response["output"]

if "session_id" not in st.session_state:
    session_id = uuid.uuid4()
    st.session_state.session_id = str(session_id)

if "messages" not in st.session_state:
    st.session_state.messages = []

# Display chat messages from history on app rerun
for message in st.session_state.messages:
    with st.chat_message(message["role"]):
        st.markdown(message["content"])

# Accept user input
if prompt := st.chat_input("What is up?"):
    st.session_state.messages.append({"role": "user", "content": prompt})
    with st.chat_message("user"):
        st.markdown(prompt)

    response = get_response(prompt)
    with st.chat_message("assistant"):
        st.markdown(response)
    st.session_state.messages.append({"role": "assistant", "content": response})