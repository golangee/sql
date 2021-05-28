-- A big example that is used for some tests.

CREATE TABLE IF NOT EXISTS Artist (
    Id INT PRIMARY KEY,
    Name VARCHAR(255) NOT NULL UNIQUE,
    BirthYear INT NOT NULL
);

CREATE TABLE Song (
    Id INT PRIMARY KEY,
    Name VARCHAR(255) NOT NULL,
    Album INT,
    FOREIGN KEY (Album) REFERENCES Album(Id)
);

-- With this table multiple artists can work on the same song.
CREATE TABLE WorkedOn (
    Artist INT NOT NULL,
    Song INT NOT NULL,
    CONSTRAINT Wrote FOREIGN KEY (Artist) REFERENCES Artist(Id),
    CONSTRAINT WrittenBy FOREIGN KEY (Song) REFERENCES Song(Id)
);

CREATE TABLE Album (
    Id INT PRIMARY KEY,
    Name VARCHAR(255),
    Year INT DEFAULT 2000
);

CREATE TABLE Publisher (
    Id INT PRIMARY KEY,
    Uuid INT,
    Year INT,
    KEY k_uuid (Uuid),
    KEY (Year)
);
