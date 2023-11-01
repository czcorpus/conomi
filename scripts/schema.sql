CREATE TABLE reports (
    id int(11) NOT NULL AUTO_INCREMENT,
    app varchar(50) NOT NULL,
    instance varchar(50) NOT NULL,
    level varchar(50) NOT NULL,
    subject text NOT NULL,
    body text NOT NULL,
    created datetime DEFAULT NOW() NOT NULL,
    resolved_by_user_id int DEFAULT NULL,
    PRIMARY KEY (id)
)