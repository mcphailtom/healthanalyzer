CREATE VIRTUAL TABLE IF NOT EXISTS vec_embeddings USING vec0(
    source_id   TEXT,
    source_type TEXT,
    embedding   float[1536]
);
