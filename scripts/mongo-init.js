// MongoDB initialization script
db = db.getSiblingDB('nishiki');

// Create collections
db.createCollection('collections');
db.createCollection('containers');

// Collections indexes
db.collections.createIndex({ "user_id": 1 });
db.collections.createIndex({ "group_id": 1 });

// Containers indexes
db.containers.createIndex({ "group_id": 1 });
db.containers.createIndex({ "user_id": 1 });
db.containers.createIndex({ "category_id": 1 });
db.containers.createIndex({ "collection_id": 1 });
db.containers.createIndex({ "parent_container_id": 1 });
db.containers.createIndex({ "objects.id": 1 });

print('Nishiki database initialized successfully');
