cd ..
cd 1.database_mysql
docker-compose up -d
//
cd 2.2.golang
go run main.go

//
cd 3.admin_cloth_nextjs13
npm run dev
///
cd 4.nginx
//
docker-compose up -d --no-build
docker-compose down
docker-compose up --build
docker build -t mysqldatabase:0.0.2 .
docker ps check continer running
docker volume ls
docker volume prune -f
docker system prune -a
docker volume rm 1database_mysql_mysql-data

# //

cd 1.1.postgres
docker build -t mypostgresAI:0.0.1 .

psql -h localhost -d restaurant -U restaurant -W

# psql -U restaurant -d restaurant

\dt : list all table
\d users
\d order_items

--
--------------------------------------------- show data ---------------------------------------
SELECT \_ FROM blog*cats; show data
SELECT FROM orders; show data
SELECT * FROM blog*comments; show data
SELECT * FROM users; show data
SELECT \_ FROM blog_posts; show data
SELECT \* FROM blog_cats; show data
SELECT \* FROM blog_sub_des; in detail
--------------------------------------------- show model ---------------------------------------
\d blog_posts
\d blog_sub_des
--

--------------------------------------------- delete ---------------------------------------
blog_comments, blog_posts, blog_sub_des, category_blogs

DROP TABLE blog_categoryasfasdfs CASCADE;
DROP TABLE blog_posts CASCADE;
DROP TABLE blog_sub_des CASCADE;
DROP TABLE blog_comments CASCADE;
DROP TABLE posts CASCADE;
------------------------- general list all table and check schema---------------------
\dt : list all table

\d blog_posts

-------------------------------------------- Verify Foreign Key Definition:---------------------
SELECT \*
FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
WHERE TABLE_NAME = 'blog_posts'
AND CONSTRAINT_NAME = 'fk_blog_cats_blog_post';

--------------------------------------------- delete data ---------------------------------------
DELETE FROM blog_posts;
DELETE FROM blog_sub_des;
UPDATE categories
SET product_id = 1
WHERE id = 1;
//------
DROP TABLE IF EXISTS products;

//

SELECT \* FROM orders
WHERE id = 2;
SELECT FROM users; show data
//
DROP TABLE "order_items";
DROP TABLE "orders";
DELETE FROM tags;
DELETE FROM order_items
WHERE id IN (25);
DROP TABLE users CASCADE;
DROP TABLE order_items CASCADE;
DROP TABLE orders CASCADE;

DROP TABLE tokens CASCADE;
//
CREATE EXTENSION vector1;
order_Items
//
\q

CREATE TABLE order_items (
Id UUID PRIMARY KEY,
OrderId UUID,
ProductId UUID,
CustomerId INT,
Quantity INT,
Total FLOAT
);

<!-- SELECT * FROM billboards; -->

INSERT INTO Bard_Response_Model (response) VALUES ('This is a Bard response.');

<!-- SELECT * FROM users; -->

# cd ..

<!--  -->

cd 2.2.golang
docker build -t go:0.0.1 .
//

cd 2.2.golang
go run main.go

go mod init golang-crud-gin

go clean

go mod tidy
http://localhost:8888/docs/index.html

cd 2.1.fastapi_ai
uvicorn main:app --host 127.0.0.1 --port 8000
python3 -m venv env
source env/bin/activate
pip install fastapi sqlalchemy psycoqg2-binary

pip install psycopg2-binary

# //

//

cd 3.admin_cloth_nextjs13
npm run dev
npm install typewriter-effect

# //

python3 --version
pip3 install google-bard-api

//
cd 4.nginx
docker build -t nginxbackend:0.0.1 .
//

cd fastapi-clean-architecture-main
uvicorn app.main:app --host 127.0.0.1 --port 9001

http://127.0.0.1:9001/docs -- for swageer
pip install dependency-injector
pip install psycopg2-binary
pip install alembic
alembic init alembic2
link database
alembic upgrade head
pip install pgvector

alembic revision --autogenerate -m "Initial migration"
alembic -x revision --autogenerate -m "revision_name"
//

cd fastapi-clean-architecture-maintest
uvicorn app.main:app --host 127.0.0.1 --port 9000
//

cd 5.redis
docker build -t myredis:0.0.1 .

# cd 4.fe_cloth_nextjs13

# npm run dev

# rm -rf node_modules

# rm package-lock.json # if you're using npm

# rm yarn.lock # if you're using yarn

# npm cache clean --force

# npm install

# cloudinary

# upload preset: atk9scxn

# foldername: cloth1

# api: CLOUDINARY_URL=cloudinary://753287843977748:GZSqk_zGNYvXxPHIh86VWtye4xU@dirs6avbc

# name: dirs6avbc

# apikey: 753287843977748

# api secret: GZSqk_zGNYvXxPHIh86VWtye4xU

change tags to uesrs
and Tags to Users
mysql -u root -p
myrootpassword

SHOW DATABASES;
USE mydatabase;
SHOW TABLES;
SELECT \* FROM products;
DESCRIBE billboards;

SELECT id, name, email FROM users;

# "Email": "userasdfasdf@email.com",

# "Password": "yourSecurePassword",

# "Image": "path/to/image.jpg",

# "EmailVerified": null,

# "FavoriteIds": ["fav1", "fav2"],

# "Role": "client"

# {

# "name": "fas1aad454fsdf",

# "email": "sd2sxample@asdfasdfsadf.com",

# "password": "123456",

# "image": "https://example.com/image.jpg",

# "emailVerified": "2023-08-28T12:00:00Z",

# "favoriteIds": ["123", "456"],

# "role": "client"

# }

# {

# "name": "sdfgsdf1gdfg",

# "email": "esawzq1d@sdfgdsfg.com",

# "password": "123456",

# "image": "https://example.com/image.jpg",

# "emailVerified": "2023-08-28T12:00:00Z",

# "favoriteIds": ["123", "456"],

# "role": "client"

# }

# # change users to billboards

# # and Users to Billboards

# # import { toast } from "react-hot-toast"; -->

// git hub

 <!-- fast api -->

cd fastapitest
uvicorn main:app --host 127.0.0.1 --port 9000
uvicorn main:app --reload
pip3 list
python3 -m vevn fastapienv
pip install --upgrade sqlmodel
source fastapienv/bin/activate
deactive - deactive enviroment
pip install fastapi, "uvicorn[standdard]"
pip list to check list
pip install Jinja2
pip install "python-jose[cryptography"]

<!--  -->
   <!-- <Image
        src="http://127.0.0.1:8888/uploads/s1 orange.jpg"
        alt=""
        width={100} // Specify the width in pixels
        height={100} // Specify the height in pixels
        // layout="responsive" // Choose an appropriate layout, e.g., "responsive"
        // className="aspect-square object-cover rounded-md"
      /> -->

option z to align the code contend
npm install react-quill
npm install tiptap

--

Checking Git Hub history

git log

git checkout -b new-brand-due-to-dart-light-bg eda5ca118dc62b592354254a976abc05f6c53277

git checkout master

--------------------------------------------- ubuntu ---------------------------------------
cd 1.ubuntu
docker build -t ubuntu-server .
docker run -it --privileged ubuntu-server bash
apt-get update && apt-get upgrade
sudo apt update
sudo apt upgrade

docker run -it ubuntu-server bash
FROM ubuntu:22.04

# Update the system efficiently (non-interactive)

RUN apt-get update && apt-get upgrade -y --no-install-recommends

1. login as root user

-------------- update appp 2. sudo apt update
sudo apt upgrade

3. cat etc/apt/sources.list
   u// ----------------- íntall ubuntu
4. sudo apt install nginx
5. sudo nginx -v
   ------------------------ ufw
   sudo apt install ufw

6. sudo ufw app list (tuong lua chan nginx)

sudo ufw allow 'Nginx HTTP'
ufw status
sudo systemctl start nginx
sudo dpkg -l | grep systemd
sudo apt-get install systemd
sudo service nginx start
service nginx stop
sudo service status nginx

cd /usr/local/nginx/conf

cd /etc/nginx

sudo nano nginx.conf
sudo vim /nginx.conf
emacs /nginx.conf
--------------------------------------------- docker ---------------------------------------

docker ps - check running docker
docker images
docker rmi my-ubuntu-server

//-------------------------------- server

ssh root@14.225.44.227

ssh sammy@14.225.44.227
n2BT98vVDV4Q2pcQ9fSJeC3
su -
getent passwd
su - sammy
-------------------- PostgreSQL on Ubuntu 20.04 -------------
\conninfo : You are connected to database "sammy" as user "sammy" via socket in "/var/run/postgresql" at port "5432".
sudo -i -u postgres
psql

\q
createuser --interactive
sammy
createdb sammy
psql -l
//---------------------------

sudo apt install postgresql postgresql-contrib

sudo systemctl status postgresql
psql --version
which psql

sudo -u postgres psql
\q

sudo -u postgres createuser --interactive

sudo adduser sammy

pass: 123456789
full name sammy
Enter the new value, or press ENTER for the default
Full Name []: sammy
Room Number []: 1
Work Phone []: 123456789
Home Phone []: 123456789
Other []:
Is the information correct? [Y/n] y

getent passwd | awk -F: '$3 >= 1000 && $3 < 60000 {print $1}' | wc -l
1
root@ubuntu-2vcpu-2gb-1-0xe1q:~#
getent passwd | wc -l

getent passwd
sudo whoami
sudo adduser sammy

sudo -i -u sammy
psql

psql -h localhost -U sammy

createdb database1
createuser database1

psql -h localhost -U database1

sudo -u postgres createdb sammy

----------------------------------- docker -------------------------------

docker images
docker-compose build go_app
2.2.golang/Dockerfile
docker-compose up -d

--------------------- nginx --------------------------
/opt/homebrew/bin/nginx -c /Users/monghoaivu/Desktop/code/english\ app\ main/english_app-main/4.nginx/config/nginx/default.conf
/opt/homebrew/bin/nginx -c /Users/monghoaivu/Desktop/code/english\ app\ main/english_app-main/4.nginx/config/nginx/default.conf
nginx -v
cd 4.nginx
nginx -t
sudo nginx

lsof -i :8080
which nginx

sudo nginx -s stop
ps -ef | grep nginx

/opt/homebrew/bin/nginx

sudo /Users/monghoaivu/Desktop/code/english app main/english_app-main/4.nginx/config/nginx

4.nginx/config/nginx

sudo /Users/monghoaivu/Desktop/code/english\ app\ main/english_app-main/4.nginx/config/nginx

nginx -s reload

postgres

buid image:

cd postgres
docker build -t mypostgres_ai:0.0.1 .
docker-compose up -d
docker-compose ps