CREATE TABLE conomi_report_group (
    id int(11) NOT NULL AUTO_INCREMENT,
    app varchar(50) NOT NULL,
    instance varchar(50),
    tag varchar(100),
    created datetime DEFAULT NOW() NOT NULL,
    resolved_by_user_id int DEFAULT NULL,
    severity varchar(50) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE conomi_report (
    id int(11) NOT NULL AUTO_INCREMENT,
    report_group_id int(11) NOT NULL REFERENCES conomi_report_group(id),
    severity varchar(50) NOT NULL,
    subject text NOT NULL,
    body text NOT NULL,
    args json,
    created datetime DEFAULT NOW() NOT NULL,
    PRIMARY KEY (id)
);