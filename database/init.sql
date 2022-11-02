CREATE TABLE IF NOT EXISTS service_user (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS main_account (
    id INT PRIMARY KEY AUTO_INCREMENT,
    balance INT UNSIGNED,
    service_user_id INT,
    FOREIGN KEY (service_user_id) REFERENCES service_user(id)
);

CREATE TABLE IF NOT EXISTS reserve_account (
    id INT PRIMARY KEY AUTO_INCREMENT,
    balance INT UNSIGNED,
    service_user_id INT,
    FOREIGN KEY (service_user_id) REFERENCES service_user(id)
);

CREATE TABLE IF NOT EXISTS reservation (
    id INT PRIMARY KEY AUTO_INCREMENT,
    service_id INT NOT NULL,
    order_id INT NOT NULL,
    service_user_id INT,
    amount INT,
    FOREIGN KEY (service_user_id) REFERENCES service_user(id)
);

-- INSERT INTO user_account (amount) VALUES (0);
