CREATE TABLE comics_keywords (
    comics_id INT NOT NULL,
    word TEXT NOT NULL,
    
    PRIMARY KEY(comics_id, word),
    FOREIGN KEY (comics_id) REFERENCES comics(id)
);