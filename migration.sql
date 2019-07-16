create database if not exists address_contact;

use address_contact;

create table if not exists address_contact (
    id bigint unsigned not null auto_increment,
    name varchar(255) not null,
    address varchar(255) not null,
    email varchar(255) not null,
    primary key(id)
);
