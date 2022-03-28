DROP DATABASE IF EXISTS artDb;

CREATE DATABASE artDb;

USE artDb;

SET CHARACTER SET utf8mb4;

/* +----------------- UTIL ----------------+ */

CREATE TABLE NationTable (
	Id INT NOT NULL AUTO_INCREMENT,
	Name VARCHAR(30) NOT NULL,
	PRIMARY KEY ( Id )
);

/* +----------------- ART TABLES ----------------+ */

/* artwork table, saves basic data for artworks */
CREATE TABLE Artworks (
	Id INT NOT NULL AUTO_INCREMENT,
	OriginalTitle VARCHAR(200) NOT NULL,
	Title VARCHAR(200), /* if the title is in english, them title will be null */
	YearOfCreation INT NOT NULL,
	Description TEXT,
	Likes INT NOT NULL DEFAULT 0,
	AudioCommentary BLOB, /* small audio commentary */
	Owner VARCHAR(100), /* describe current owner, might be null */
	BorrowedTo VARCHAR(100), /* describe current position of the artwork, might be null */
	PRIMARY KEY ( Id )
);

/* artist table, saves all data related to Artists */
CREATE TABLE Artists (
	Id INT NOT NULL AUTO_INCREMENT,
	Name VARCHAR(30) NOT NULL,
	SecondName VARCHAR(100), /* second names, migh be null */
	Surname VARCHAR(75) NOT NULL,
	DateOfBirth DATE NOT NULL,
	Nationality INT NOT NULL,
	Description TEXT, /* small description */
	PRIMARY KEY ( Id ),
	FOREIGN KEY ( Nationality ) REFERENCES NationTable( Id )
);

/*
creator table:
Since there may be multiple creators we need a separate table defining all the artist who worked on a artwork
*/
CREATE TABLE CreatedBy (
	Id INT NOT NULL AUTO_INCREMENT,
	ArtworkId INT NOT NULL, /* foreign key, defines artwork */
	ArtistId INT NOT NULL, /* foreing key, define the artists who created the artwork */
	PRIMARY KEY ( Id ),
	FOREIGN KEY ( ArtworkId ) REFERENCES Artworks( Id ),
	FOREIGN KEY ( ArtistId ) REFERENCES Artists( Id )
);

/* artwork picture table */
CREATE TABLE ArtworkPicture (
	Id INT NOT NULL AUTO_INCREMENT,
	ArtworkId INT NOT NULL,
	PicturePath VARCHAR(30) NOT NULL,
	PRIMARY KEY ( Id ),
	FOREIGN KEY ( ArtworkId ) REFERENCES Artists( Id )
);

/* comment table */
CREATE TABLE Comments (
	Id INT NOT NULL AUTO_INCREMENT,
	ArtworkId INT NOT NULL,
	Username VARCHAR(30) NOT NULL,
	Comment TEXT NOT NULL,
	PRIMARY KEY ( Id ),
	FOREIGN KEY ( ArtworkId ) REFERENCES Artists( Id )
);

/* +----------------- TABLES POPULATION -------------------+ */

INSERT INTO NationTable
    ( Name )
VALUES
	( 'France' ),
	( 'Italy' ),
	( 'United States' )
;
INSERT INTO Artists
	( Name, SecondName, Surname, DateOfBirth, Nationality, Description )
VALUES
    ( "Vincent", "Willem", "van Gogh", "1853-03-30", 1, "Vincent Willem van Gogh was a Dutch Post-Impressionist painter who posthumously became one of the most famous and influential figures in Western art history. In a decade, he created about 2,100 artworks, including around 860 oil paintings, most of which date from the last two years of his life." ),
	( "Leonardo", "di ser Piero", "Da Vinci", "1452-04-14", 2, "Leonardo di ser Piero da Vinci was an Italian polymath of the High Renaissance who was active as a painter, draughtsman, engineer, scientist, theorist, sculptor and architect."),
	( "Jackson", "Paul", "Pollock", "1912-01-28", 3, "Paul Jackson Pollock was an American painter and a major figure in the abstract expressionist movement. He was widely noticed for his 'drip technique' of pouring or splashing liquid household paint onto a horizontal surface, enabling him to view and paint his canvases from all angles." )
;

INSERT INTO Artworks
	( OriginalTitle, Title, YearOfCreation, Description, Owner, BorrowedTo )
VALUES
    ( 'Champ de blé aux corbeaux', 'Wheatfield with Crows', 1890, "Wheatfield with Crows is a July 1890 painting by Vincent van Gogh. It has been cited by several critics as one of his greatest works. It is commonly stated that this was van Gogh's final painting. However, art historians are uncertain as to which painting was van Gogh's last, as no clear historical records exist.", 'Netherlands state', 'Van Gogh Museum' ),
	( 'La Belle Ferronnière', 'Portrait of an Unknown Woman', 1490, "La Belle Ferronnière is a portrait of a lady, usually attributed to Leonardo da Vinci, in the Louvre. It is also known as Portrait of an Unknown Woman. The painting's title, applied as early as the seventeenth century, identifying the sitter as the wife or daughter of an ironmonger, was said to be discreetly alluding to a reputed mistress of Francis I of France, married to a certain Le Ferron.", 'Musée du Louvre', NULL ),
	( 'Blue Poles', NULL, 1952, "Blue Poles, also known as Number 11, 1952 is an abstract expressionist painting by American artist Jackson Pollock. It was purchased amid controversy by the National Gallery of Australia in 1973 and today remains one of the gallery's major paintings.", 'Blue Poles', NULL )
;

INSERT INTO CreatedBy
	( ArtworkId, ArtistId )
VALUES
	( 1, 1 ),
	( 2, 2 ),
	( 3, 3 )
;

INSERT INTO ArtworkPicture
    ( ArtworkId, PicturePath )
VALUES
    ( 1, "wfc1.jpeg" ),
    ( 1, "wfc1.jpeg" ),
    ( 1, "wfc3.jpeg" ),
    ( 2, "lbf1.jpeg" ),
    ( 2, "lbf2.jpeg" ),
    ( 3, "bp1.jpeg" )
;