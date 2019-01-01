CREATE EXTENSION "citext";

CREATE TABLE "user" (
    "id" INTEGER
        CONSTRAINT "user_id_pk" PRIMARY KEY,
    "nickname" CITEXT
        CONSTRAINT "user_nickname_not_null" NOT NULL
        CONSTRAINT "user_nickname_unique" UNIQUE,
    "fullname" TEXT
        CONSTRAINT "user_fullname_not_null" NOT NULL,
    "email" TEXT
        CONSTRAINT "user_email_not_null" NOT NULL
        CONSTRAINT "user_email_unique" UNIQUE,
    "about" TEXT
        CONSTRAINT "user_about_nullable" NULL
);
