---
name: create-database-migration
description: Create SQL database migration scripts for new features and engineering improvements. Use this skill to understand how to create database migrations in a consistent manner across the project, following our engineering guidelines. 
---

# Create Database Migration

## Overview 
* We use Postrgres as our sql database. 
* We use Goose to manage database migrations. 
* Migration scripts are found in the `./backend/migrations/` directory. 

## Migration Script Instructions 
* For every table, include a `deleted_at` field. We will use soft deletes for everything at the application layer. 
* For every table, include an `updated_at` field that will refect the last time that a record was updated. 
* Use UUIDs for primary keys. 
* Use `TEXT` fields for strings instead of `VARCHAR(n)` 