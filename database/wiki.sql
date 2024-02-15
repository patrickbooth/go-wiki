CREATE TABLE IF NOT EXISTS user (
  userID        SMALLINT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY,
  realName      TINYTEXT NOT NULL,
  userName      TINYTEXT NOT NULL,
  createdDate   DATETIME NOT NULL
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS pages (
  pageID        SMALLINT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY,
  title         TINYTEXT NOT NULL,
  body          TEXT,
  createdDate   DATETIME NOT NULL,
  updatedDate   DATETIME NOT NULL,
  userID        SMALLINT UNSIGNED NOT NULL,
  updatedBy     SMALLINT,
  CONSTRAINT `fk_page_user`
    FOREIGN KEY (userID) REFERENCES user (userID)
    ON DELETE RESTRICT
    ON UPDATE RESTRICT
) ENGINE = InnoDB;

INSERT INTO user
  (realName, userName, createdDate)
VALUES
  ('Poetic Gopher', 'poetic-gopher', NOW());

INSERT INTO pages
  (title, body, createdDate, updatedDate, userID, updatedBy)
VALUES
  ('Lorem Ipsum', 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Praesent dictum, arcu eget consectetur interdum, risus magna viverra nibh, at placerat lacus est in dui. Aenean quis nisi id justo laoreet aliquam. Nunc a lectus interdum, rutrum massa quis, commodo nisl. Vestibulum ullamcorper, nunc id rutrum tempus, mi urna accumsan ex, ac faucibus elit lorem dictum nulla. Etiam dapibus, nulla non consectetur imperdiet, lacus lacus viverra lacus, dignissim gravida nulla odio eget libero. Suspendisse ac ex fermentum, vestibulum lorem eget, porta sapien. Vestibulum tempor enim vel elit egestas, eu hendrerit libero imperdiet. ', NOW(), NOW(), LAST_INSERT_ID(), LAST_INSERT_ID());

