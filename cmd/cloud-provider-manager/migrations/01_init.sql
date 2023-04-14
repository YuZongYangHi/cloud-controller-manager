CREATE DATABASE `cloud_privoder` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `cloud_privoder`;

CREATE TABLE `loadbalances`(
    `id` bigint(20) NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `cluster` VARCHAR(255) NOT NULL,
    `ip` VARCHAR(255) NOT NULL,
    `carriers` int(10) NOT NULL,
    `status` int(10) NOT NULL DEFAULT 0,
    `cidr` varchar(255) NOT NULL,
    `created_at` datetime(6) NOT NULL,
    `updated_at` datetime(6) NOT NULL
);

INSERT INTO
    loadbalances(
    cluster,
    ip,
    carriers,
    cidr,
    created_at,
    updated_at
)
VALUES
    (
        'cdcm21',
        '172.28.205.200',
        0,
        '172.28.205.200/29',
        '2023-04-13 00:00:00',
        '2023-04-13 00:00:00'
    ),
    (
        'cdcm21',
        '172.28.205.201',
        0,
        '172.28.205.200/29',
        '2023-04-13 00:00:00',
        '2023-04-13 00:00:00'
    ),
    (
        'cdcm21',
        '172.28.205.202',
        0,
        '172.28.205.200/29',
        '2023-04-13 00:00:00',
        '2023-04-13 00:00:00'
    ),
    (
        'cdcm21',
        '172.28.205.203',
        0,
        '172.28.205.200/29',
        '2023-04-13 00:00:00',
        '2023-04-13 00:00:00'
    ),
    (
        'cdcm21',
        '172.28.205.204',
        0,
        '172.28.205.200/29',
        '2023-04-13 00:00:00',
        '2023-04-13 00:00:00'
    ),
    (
        'cdcm21',
        '172.28.205.205',
        0,
        '172.28.205.200/29',
        '2023-04-13 00:00:00',
        '2023-04-13 00:00:00'
    ),
    (
        'cdcm21',
        '172.28.205.206',
        0,
        '172.28.205.200/29',
        '2023-04-13 00:00:00',
        '2023-04-13 00:00:00'
    );