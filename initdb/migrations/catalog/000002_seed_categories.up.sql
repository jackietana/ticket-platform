INSERT INTO categories (id, name) VALUES
    (1, 'sport'),
    (2, 'theater'),
    (3, 'standup'),
    (4, 'concert'),
    (5, 'esports'),
    (6, 'golf'),
    (7, 'rally'),
    (8, 'marathon'),
    (9, 'hackathon'),
    (10, 'holiday')
ON CONFLICT (id) DO NOTHING;
