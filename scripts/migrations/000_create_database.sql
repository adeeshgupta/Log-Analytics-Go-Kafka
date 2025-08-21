-- Database Creation Migration
-- This script creates the database if it doesn't exist
-- This should be the first migration to run

-- Create database if it doesn't exist
CREATE DATABASE IF NOT EXISTS log_analytics CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Use the database
USE log_analytics;

-- Database creation completed successfully 
