CREATE DATABASE IF NOT EXISTS mtg_inventory;

USE mtg_inventory;

CREATE TABLE IF NOT EXISTS users (
	username VARCHAR(256) NOT NULL PRIMARY KEY,
	email VARCHAR(256)
);

CREATE TABLE IF NOT EXISTS singles (
	id INT NOT NULL PRIMARY KEY,
	oracle_id VARCHAR(256) NOT NULL,
	expansion VARCHAR(256),
	language VARCHAR(256),
	owner VARCHAR(256) NOT NULL,
	holder VARCHAR(256) NOT NULL,
	FOREIGN KEY (owner) REFERENCES users(username),
	FOREIGN KEY (holder) REFERENCES users(username)
);

CREATE TABLE IF NOT EXISTS requests (
	id INT NOT NULL PRIMARY KEY,
	requestor VARCHAR(256) NOT NULL,
	created	DATETIME NOT NULL,
	FOREIGN KEY (requestor) REFERENCES users(username)
);

CREATE TABLE IF NOT EXISTS requested_cards (
	request_id INT NOT NULL,
	oracle_id VARCHAR(256) NOT NULL,
	quantity INT NOT NULL,
	UNIQUE(request_id, oracle_id),
	FOREIGN KEY (request_id) REFERENCES requests(id)
);
