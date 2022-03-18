DROP DATABASE IF EXISTS instanTex_db;

CREATE DATABASE instanTex_db;

USE instanTex_db;

CREATE TABLE Users (
    id INT NOT NULL AUTO_INCREMENT,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    date_of_join DATE NOT NULL,
    salt INT NOT NULL,
    pHash CHAR(64) NOT NULL,
	 last_login TIMESTAMP NOT NULL,
    PRIMARY KEY ( id )
);

CREATE TABLE FriendRequests (
    id INT NOT NULL AUTO_INCREMENT,
    senderId INT NOT NULL,
    reciverId INT NOT NULL,
    PRIMARY KEY ( id )
);

CREATE TABLE ConversationName (
	id INT NOT NULL AUTO_INCREMENT,
	name VARCHAR(30) NOT NULL,
	description TEXT,
	date_of_creation DATE NOT NULL DEFAULT CURRENT_DATE,
	PRIMARY KEY ( id )
);

CREATE TABLE Conversations (
    id INT NOT NULL,
    participantId INT NOT NULL,
    date_of_join DATE NOT NULL DEFAULT CURRENT_DATE,
    PRIMARY KEY ( id, participantId )
);

CREATE TABLE Token (
    userid INT NOT NULL,
    accessToken CHAR(64) NOT NULL,
    act_expt INT NOT NULL,
    refreshToken CHAR(64) NOT NULL,
	rft_expt INT NOT NULL,
    PRIMARY KEY ( userid )
);

CREATE TABLE AccessToken (
	userid INT NOT NULL,
	accessToken CHAR(64) NOT NULL,
	act_expt INT NOT NULL,
	PRIMARY KEY ( userid )
);

CREATE TABLE RefreshToken (
	Id INT AUTO_INCREMENT
	userid INT NOT NULL,
	refreshToken CHAR(64) NOT NULL,
	rft_expt INT NOT NULL,
	PRIMARY KEY ( userid )
);


CREATE TABLE PwOtp (
	userId INT NOT NULL,
	otp CHAR(32) NOT NULL,
	expt INT NOT NULL,
	PRIMARY KEY (userId)
);

CREATE TABLE LoginState (
    idstring CHAR(15) NOT NULL,
    userEmail VARCHAR(45),
    PRIMARY KEY ( idstring )
);


INSERT INTO Users ( username, email, date_of_join, salt, pHash, last_login ) VALUES ( "pima", "pippo.mario@gimelli.com", CURRENT_DATE(), 66858, "ac7c27a867f92dbecc637a14afae8657f2c2a65eb47faeb3a6cadcad21c17da0", CURRENT_TIMESTAMP() );

INSERT INTO Users ( username, email, date_of_join ,salt, pHash, last_login) VALUES ( "taurone", "taurone.mario@gimelli.com", CURRENT_DATE(), 66858, "ac7c27a867f92dbecc637a14afae8657f2c2a65eb47faeb3a6cadcad21c17da0", CURRENT_TIMESTAMP());

INSERT INTO ConversationName ( name, description  ) VALUES ( "gruppo tennici (senza taurone)", "tutti i miei amici odiano taurone" );

INSERT INTO Conversations ( id, participantId ) VALUES ( 1, 1 );

INSERT INTO ConversationName ( name, description  ) VALUES ( "gruppo tennici", "taurone trovati un lavoro" );

INSERT INTO Conversations ( id, participantId ) VALUES ( 2, 1 );

INSERT INTO Conversations ( id, participantId ) VALUES ( 2, 2 );

CREATE TABLE MessageTable0 (
    id INT NOT NULL AUTO_INCREMENT,
    senderId INT NOT NULL,
    sendDate INT NOT NULL,
    messageText VARCHAR(1000) NOT NULL,
    attachment VARCHAR(100),
    PRIMARY KEY ( id )
);

INSERT INTO MessageTable0 ( senderId, sendDate,  messageText ) VALUES ( 0, UNIX_TIMESTAMP(), "taurone aiudo mi si è rotto lu picci" );
