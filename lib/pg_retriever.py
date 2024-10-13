
from typing import List

from langchain_core.callbacks import CallbackManagerForRetrieverRun
from langchain_core.documents import Document
from langchain_core.retrievers import BaseRetriever

from langchain_google_vertexai import VertexAIEmbeddings

import psycopg2
from pgvector.psycopg2 import register_vector

class CourseContentRetriever(BaseRetriever):
    """Retriever to find relevant course content based on the
    query provided."""

    embeddings_service: VertexAIEmbeddings    
    similarity_threshold: float
    num_matches: int
    conn_str: str

    def _get_relevant_documents(
            self, query: str, *, run_manager: CallbackManagerForRetrieverRun
        ) -> List[Document]:
        conn = psycopg2.connect(self.conn_str)
        register_vector(conn)

        qe = self.embeddings_service.embed_query(query)

        with conn.cursor() as cur:
            cur.execute(
                """
                        WITH vector_matches AS (
                        SELECT id, content, 1 - (embedding <=> %s::vector) AS similarity
                        FROM course_content_embeddings
                        WHERE 1 - (embedding <=> %s::vector) > %s
                        ORDER BY similarity DESC
                        LIMIT %s
                        )
                        SELECT cc.id as id, cc.title as title, 
                            vm.content as content, 
                            vm.similarity as similarity 
                        FROM course_contents cc
                        LEFT JOIN vector_matches vm ON cc.id = vm.id;
                """,
                (qe, qe, self.similarity_threshold, self.num_matches)
            )
            results = cur.fetchall()

        conn.close()

        if not results:
            return []
        
        return [
            Document(
                page_content=r[2],
                metadata={
                    "id": r[0],
                    "title": r[1],
                    "similarity": r[3],
                }
            ) for r in results if r[2] is not None
        ]
