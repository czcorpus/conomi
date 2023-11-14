CREATE TABLE conomi_reports (
    id int(11) NOT NULL AUTO_INCREMENT,
    app varchar(50) NOT NULL,
    instance varchar(50),
    tag varchar(100),
    severity varchar(50) NOT NULL,
    subject text NOT NULL,
    body text NOT NULL,
    args json,
    created datetime DEFAULT NOW() NOT NULL,
    resolved_by_user_id int DEFAULT NULL,
    PRIMARY KEY (id)
)