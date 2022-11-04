CREATE TABLE IF NOT EXISTS service_user (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS main_account (
    id INT PRIMARY KEY AUTO_INCREMENT,
    balance DECIMAL(15,2) UNSIGNED,
    service_user_id INT,
    FOREIGN KEY (service_user_id) REFERENCES service_user(id)
);

CREATE TABLE IF NOT EXISTS reserve_account (
    id INT PRIMARY KEY AUTO_INCREMENT,
    balance DECIMAL(15,2) UNSIGNED,
    service_user_id INT,
    FOREIGN KEY (service_user_id) REFERENCES service_user(id)
);

CREATE TABLE IF NOT EXISTS reservation (
    id INT PRIMARY KEY AUTO_INCREMENT,
    service_id INT NOT NULL,
    order_id INT NOT NULL,
    service_user_id INT,
    amount DECIMAL(15,2) UNSIGNED,
    FOREIGN KEY (service_user_id) REFERENCES service_user(id)
);

CREATE TABLE IF NOT EXISTS bookkeeping (
    id INT PRIMARY KEY AUTO_INCREMENT,
    service_user_id INT,
    service_id INT NOT NULL,
    amount DECIMAL(15,2) UNSIGNED,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (service_user_id) REFERENCES service_user(id)
);

CREATE TABLE IF NOT EXISTS user_report (
    id INT PRIMARY KEY AUTO_INCREMENT,
    service_user_id INT,
    amount DECIMAL(15,2) UNSIGNED,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (service_user_id) REFERENCES service_user(id)
);

CREATE TABLE IF NOT EXISTS bookkeeping_report (
    id INT PRIMARY KEY AUTO_INCREMENT,
    hash_string VARCHAR(255),
    path_to_file VARCHAR(255)
);

INSERT INTO service_user (username) VALUES ("user1"), ("user2"), ("user3"), ("user4");
