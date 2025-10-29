// MongoDB initialization script
db = db.getSiblingDB('nishiki');

// Create collections
db.createCollection('containers');

db.containers.createIndex({ "group_id": 1 });
db.containers.createIndex({ "user_id": 1 });
db.containers.createIndex({ "category": 1 });

print('Nishiki database initialized successfully');
