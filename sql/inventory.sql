CREATE DATABASE IF NOT EXISTS mtg_inventory;

USE mtg_inventory;

CREATE TABLE IF NOT EXISTS users (
	id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
	slack_id VARCHAR(256) NOT NULL,
	UNIQUE (slack_id)
);

CREATE TABLE IF NOT EXISTS cards (
	quantity INT NOT NULL,
	name VARCHAR(256) NOT NULL,
	oracle_id VARCHAR(256) NOT NULL,
	scryfall_id VARCHAR(256) NOT NULL,
	foil BOOLEAN,
	owner INT NOT NULL,
	keeper INT NOT NULL,
	UNIQUE (scryfall_id, foil, owner, keeper),
	FOREIGN KEY (owner) REFERENCES users(id),
	FOREIGN KEY (keeper) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS requests (
	id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
	requestor INT NOT NULL,
	opened DATETIME NOT NULL,
	closed DATETIME,
	FOREIGN KEY (requestor) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS requested_cards (
	request_id INT NOT NULL,
	name VARCHAR(256) NOT NULL,
	oracle_id VARCHAR(256) NOT NULL,
	quantity INT NOT NULL,
	UNIQUE (request_id, oracle_id),
	FOREIGN KEY (request_id) REFERENCES requests(id)
);

CREATE TABLE IF NOT EXISTS transfers (
	id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
	request_id INT,
	to_user INT NOT NULL,
	from_user INT NOT NULL,
	opened DATETIME NOT NULL,
	closed DATETIME,
	FOREIGN KEY (to_user) REFERENCES users(id),
	FOREIGN KEY (from_user) REFERENCES users(id),
	FOREIGN KEY (request_id) REFERENCES requests(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS transferred_cards (
	transfer_id INT NOT NULL,
	quantity INT NOT NULL,
	name VARCHAR(256) NOT NULL,
	scryfall_id VARCHAR(256) NOT NULL,
	foil BOOLEAN,
	owner INT NOT NULL,
	UNIQUE (transfer_id, scryfall_id, foil, owner),
	FOREIGN KEY (transfer_id) REFERENCES transfers(id) ON DELETE CASCADE,
	FOREIGN KEY (owner) REFERENCES users(id)
);
