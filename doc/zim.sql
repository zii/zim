/*
 Navicat Premium Data Transfer

 Source Server         : 86
 Source Server Type    : MySQL
 Source Server Version : 50736
 Source Host           : 10.10.10.86:3306
 Source Schema         : zim

 Target Server Type    : MySQL
 Target Server Version : 50736
 File Encoding         : 65001

 Date: 21/09/2022 10:42:08
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for channel_msgbox
-- ----------------------------
DROP TABLE IF EXISTS `channel_msgbox`;
CREATE TABLE `channel_msgbox` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `chat_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '超级群id',
  `msg_id` bigint(20) NOT NULL DEFAULT '0',
  `msg_blob` blob,
  `from_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `created_at` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq` (`chat_id`,`msg_id`) USING BTREE,
  KEY `chat_id` (`chat_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=COMPRESSED;

-- ----------------------------
-- Table structure for chat
-- ----------------------------
DROP TABLE IF EXISTS `chat`;
CREATE TABLE `chat` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `chat_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `owner_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `type` smallint(6) NOT NULL DEFAULT '0',
  `title` varchar(100) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `about` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `photo` varchar(120) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `maxp` int(11) NOT NULL DEFAULT '0',
  `muted` tinyint(4) NOT NULL DEFAULT '0',
  `deleted` tinyint(4) NOT NULL DEFAULT '0',
  `created_at` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq` (`chat_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- ----------------------------
-- Table structure for chat_member
-- ----------------------------
DROP TABLE IF EXISTS `chat_member`;
CREATE TABLE `chat_member` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `chat_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `user_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `name` varchar(50) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `role` smallint(6) NOT NULL DEFAULT '0',
  `muted` tinyint(4) NOT NULL DEFAULT '0',
  `deleted` tinyint(4) NOT NULL DEFAULT '0',
  `created_at` int(11) NOT NULL DEFAULT '0',
  `updated_at` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq` (`chat_id`,`user_id`),
  KEY `chat_id` (`chat_id`) USING BTREE,
  KEY `user_id` (`user_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- ----------------------------
-- Table structure for event
-- ----------------------------
DROP TABLE IF EXISTS `event`;
CREATE TABLE `event` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `self_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `peer_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `seq` bigint(20) NOT NULL DEFAULT '0',
  `msg_blob` blob,
  `created_at` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq` (`self_id`,`seq`) USING BTREE,
  KEY `self-peer` (`self_id`,`peer_id`) USING BTREE,
  KEY `self_id` (`self_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- ----------------------------
-- Table structure for friend
-- ----------------------------
DROP TABLE IF EXISTS `friend`;
CREATE TABLE `friend` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `peer_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `name` varchar(50) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `blocked` tinyint(4) NOT NULL DEFAULT '0',
  `created_at` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq` (`user_id`,`peer_id`) USING BTREE,
  KEY `user_id` (`user_id`) USING BTREE,
  KEY `peer_id` (`peer_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- ----------------------------
-- Table structure for friend_apply
-- ----------------------------
DROP TABLE IF EXISTS `friend_apply`;
CREATE TABLE `friend_apply` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `from_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `to_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `hash` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `greets` json DEFAULT NULL,
  `name` varchar(50) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `status` smallint(6) NOT NULL DEFAULT '0',
  `updated_at` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq` (`user_id`,`hash`) USING BTREE,
  KEY `user_id` (`user_id`) USING BTREE,
  KEY `from-to` (`from_id`,`to_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- ----------------------------
-- Table structure for message
-- ----------------------------
DROP TABLE IF EXISTS `message`;
CREATE TABLE `message` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `msg_id` bigint(20) NOT NULL DEFAULT '0',
  `msg_blob` blob,
  `from_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `to_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `created_at` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uninq` (`msg_id`) USING BTREE,
  KEY `to_id` (`to_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- ----------------------------
-- Table structure for setting
-- ----------------------------
DROP TABLE IF EXISTS `setting`;
CREATE TABLE `setting` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `k` varchar(50) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `v` text COLLATE utf8mb4_bin NOT NULL,
  `version` int(11) NOT NULL DEFAULT '0',
  `updated_at` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uniq` (`k`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `name` varchar(50) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `photo` varchar(120) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `ex` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `status` smallint(6) NOT NULL DEFAULT '0',
  `created_at` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq` (`user_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- ----------------------------
-- Table structure for user_msgbox
-- ----------------------------
DROP TABLE IF EXISTS `user_msgbox`;
CREATE TABLE `user_msgbox` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `peer_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `msg_id` bigint(20) NOT NULL DEFAULT '0',
  `from_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `to_id` varchar(40) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `created_at` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq` (`user_id`,`msg_id`) USING BTREE,
  UNIQUE KEY `uniq2` (`user_id`,`peer_id`,`msg_id`) USING BTREE,
  KEY `user_id` (`user_id`) USING BTREE,
  KEY `msg_id` (`msg_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=COMPRESSED;

SET FOREIGN_KEY_CHECKS = 1;
