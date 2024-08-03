ALTER TABLE movies
ADD CONSTRAINT movies_runtime_check CHECK (runtime >= 1 AND runtime <= 300);

ALTER TABLE movies
ADD CONSTRAINT movies_year_check CHECK (
        publish_year BETWEEN 1888 AND date_part('year', now())
    );

ALTER TABLE movies
ADD CONSTRAINT genres_length_check CHECK (
        array_length(genres, 1) BETWEEN 1 AND 5
    );