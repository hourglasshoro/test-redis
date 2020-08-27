CREATE TABLE IF NOT EXISTS messages
(
    message_id MEDIUMINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    body       VARCHAR(10000) NOT NULL,
    user_id    MEDIUMINT      NOT NULL,
    channel_id MEDIUMINT      NOT NULL,
    type       VARCHAR(10)    NOT NULL,
    created_at TIMESTAMP      NOT NULL
--      FOREIGN KEY(user_id) REFERENCES users(user_id),
--      FOREIGN KEY(channel_id) REFERENCES channels(channel_id)
);