-- 风险决策引擎数据库初始化脚本

-- 创建数据库
CREATE DATABASE IF NOT EXISTS `risk_decision` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `risk_decision`;

-- 规则配置表
CREATE TABLE IF NOT EXISTS `rule_configs` (
	`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`rule_id` VARCHAR(64) NOT NULL COMMENT '规则ID',
	`version` VARCHAR(32) NOT NULL DEFAULT '1.0' COMMENT '版本号',
	`name` VARCHAR(128) NOT NULL COMMENT '规则名称',
	`description` VARCHAR(512) DEFAULT NULL COMMENT '规则描述',
	`type` VARCHAR(32) NOT NULL DEFAULT 'BOOLEAN' COMMENT '规则类型',
	`priority` INT NOT NULL DEFAULT 100 COMMENT '优先级',
	`status` VARCHAR(32) NOT NULL DEFAULT 'ACTIVE' COMMENT '状态',
	`condition` TEXT COMMENT '条件配置JSON',
	`actions` TEXT COMMENT '动作配置JSON',
	`config_path` VARCHAR(256) DEFAULT NULL COMMENT '来源配置文件路径',
	`is_active` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否激活',
	`created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
	`updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
	`deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
	PRIMARY KEY (`id`),
	UNIQUE KEY `uk_rule_id` (`rule_id`),
	KEY `idx_status` (`status`),
	KEY `idx_is_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='规则配置表';

-- 决策记录表
CREATE TABLE IF NOT EXISTS `decision_records` (
	`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`request_id` VARCHAR(64) NOT NULL COMMENT '请求ID',
	`business_id` VARCHAR(64) DEFAULT NULL COMMENT '业务ID',
	`decision` VARCHAR(32) NOT NULL COMMENT '决策结果',
	`decision_code` VARCHAR(64) DEFAULT NULL COMMENT '决策代码',
	`decision_reason` VARCHAR(512) DEFAULT NULL COMMENT '决策原因',
	`input_data` TEXT COMMENT '输入数据JSON',
	`rule_results` TEXT COMMENT '规则结果JSON',
	`model_results` TEXT COMMENT '模型结果JSON',
	`duration_ms` BIGINT NOT NULL DEFAULT 0 COMMENT '执行耗时(ms)',
	`rule_version` VARCHAR(64) DEFAULT NULL COMMENT '规则版本',
	`session_id` VARCHAR(64) DEFAULT NULL COMMENT '会话ID',
	`created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
	`updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
	`deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
	PRIMARY KEY (`id`),
	UNIQUE KEY `uk_request_id` (`request_id`),
	KEY `idx_business_id` (`business_id`),
	KEY `idx_decision` (`decision`),
	KEY `idx_session_id` (`session_id`),
	KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='决策记录表';

-- 记录会话表
CREATE TABLE IF NOT EXISTS `recording_sessions` (
	`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`session_id` VARCHAR(64) NOT NULL COMMENT '会话ID',
	`name` VARCHAR(128) NOT NULL COMMENT '会话名称',
	`start_time` DATETIME NOT NULL COMMENT '开始时间',
	`end_time` DATETIME DEFAULT NULL COMMENT '结束时间',
	`is_active` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否活跃',
	`record_count` INT NOT NULL DEFAULT 0 COMMENT '记录数',
	`created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
	`updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
	`deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
	PRIMARY KEY (`id`),
	UNIQUE KEY `uk_session_id` (`session_id`),
	KEY `idx_is_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='记录会话表';

-- 回放会话表
CREATE TABLE IF NOT EXISTS `replay_sessions` (
	`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`session_id` VARCHAR(64) NOT NULL COMMENT '会话ID',
	`name` VARCHAR(128) NOT NULL COMMENT '会话名称',
	`recording_id` VARCHAR(64) NOT NULL COMMENT '记录会话ID',
	`config_path` VARCHAR(256) DEFAULT NULL COMMENT '规则配置路径',
	`start_time` DATETIME NOT NULL COMMENT '开始时间',
	`end_time` DATETIME DEFAULT NULL COMMENT '结束时间',
	`is_active` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否活跃',
	`report` TEXT COMMENT '回放报告JSON',
	`created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
	`updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
	`deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
	PRIMARY KEY (`id`),
	UNIQUE KEY `uk_session_id` (`session_id`),
	KEY `idx_recording_id` (`recording_id`),
	KEY `idx_is_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='回放会话表';
