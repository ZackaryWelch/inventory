// MongoDB initialization script
db = db.getSiblingDB('nishiki');

// Create collections
db.createCollection('users');
db.createCollection('groups');
db.createCollection('containers');

// Create indexes for better performance
db.users.createIndex({ "authentik_id": 1 }, { unique: true, sparse: true });
db.users.createIndex({ "email_address": 1 }, { unique: true });
db.users.createIndex({ "username": 1 }, { unique: true });

db.groups.createIndex({ "user_ids": 1 });
db.groups.createIndex({ "name": 1 });

db.containers.createIndex({ "group_id": 1 });
db.containers.createIndex({ "foods.expiry": 1 });
db.containers.createIndex({ "foods.category": 1 });

print('Nishiki database initialized successfully');