CREATE TABLE IF NOT EXISTS Log (
    Id INT PRIMARY KEY NOT NULL,
    Time INT NOT NULL, -- Useful for knowing when stuff happened!
    Author INT NOT NULL,
    Message TEXT,
    CONSTRAINT `wrote` FOREIGN KEY ("Author") REFERENCES `User`(`Id`)
);

CREATE TABLE User (
    Id INT PRIMARY KEY NOT NULL,
    Name VARCHAR(255) NOT NULL
);
