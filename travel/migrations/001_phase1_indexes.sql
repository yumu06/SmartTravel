DROP PROCEDURE IF EXISTS add_index_if_missing;

DELIMITER $$
CREATE PROCEDURE add_index_if_missing(
    IN table_name_value VARCHAR(64),
    IN index_name_value VARCHAR(64),
    IN columns_value VARCHAR(255),
    IN unique_value BOOLEAN
)
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.statistics
        WHERE table_schema = DATABASE()
          AND table_name = table_name_value
          AND index_name = index_name_value
    ) THEN
        SET @ddl = CONCAT(
            'ALTER TABLE `', table_name_value, '` ADD ',
            IF(unique_value, 'UNIQUE ', ''),
            'INDEX `', index_name_value, '` (', columns_value, ')'
        );
        PREPARE statement_value FROM @ddl;
        EXECUTE statement_value;
        DEALLOCATE PREPARE statement_value;
    END IF;
END$$
DELIMITER ;

CALL add_index_if_missing('tra_users', 'uk_open_id', '`open_id`', TRUE);
CALL add_index_if_missing('posts', 'idx_post_user_created', '`user_id`, `created_at`', FALSE);
CALL add_index_if_missing('post_comments', 'idx_comment_post_created', '`post_id`, `created_at`', FALSE);
CALL add_index_if_missing('tra_user_post_starts', 'uk_user_post', '`user_id`, `post_id`', TRUE);
CALL add_index_if_missing('tra_user_foot_starts', 'uk_user_foot', '`user_id`, `foot_id`', TRUE);
CALL add_index_if_missing('tra_foots', 'idx_foot_user_created', '`user_id`, `created_at`', FALSE);
CALL add_index_if_missing('chat_messages', 'idx_chat_from_to_created', '`from_user_id`, `to_user_id`, `created_at`', FALSE);
CALL add_index_if_missing('chat_messages', 'idx_chat_to_from_created', '`to_user_id`, `from_user_id`, `created_at`', FALSE);

DROP PROCEDURE add_index_if_missing;
