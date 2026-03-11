Create an inventory interface with the following APIs:

- Everything in nishiki-backend. Really basic group management (create groups under user, delete user groups, change names, invite users to groups via email). Groups/users can own collections of things, such food, CDs, video games, board games, or books. Currently Nishiki backend just handles food, but it also needs to handle non-food objects. See libib for important fields.
- When users login for the first time, they're Authentik uid will be linked in a MongoDB account object, and the MongoDB pk linked as an Authentik user attribute.
- Other collections, and objects within the collection, will be stored on MongoDB
- Users can add/remove objects from collections, create/rename/remove collections, share collections with other users/groups. Objects can be part of multiple collections, such as lent books and collections that represent shelves/rows. Incorporate the libib python script to sort books based off tags, keeping in mind different shelves/rows dimensions. There'll be an API to import JSON/CSV for bulk food upload  or bulk media upload, as well as an API to re-organize based on current shelf/row sizes and object sizes (books estimated based off paper, video games have some hardcoded sizes based on type, bare CDs in their own case collection. Items can also be edited from the management UI to add/remove tags, or provide dimensions to help with organizing.
- Fundamentally I'm trying to merge three apps. I'll also want to do this with spirits and cross reference cocktail recipes, but for now I'm just trying to link all the stuff I have with where I have it, and then created an automated system later to pull updates from grocery stores or manual updates. I also don't want to pay for Libib's API, a Mixel subscription, or any AWS subscription.
- Collections/rows can be tagged for what they contain. For instance, food vs board games vs books, to use with organization

In summary, expand nishiki-backend for the following APIs based off the above usecases:
- POST/GET /accounts/:id/groups
- PUT/DELETE /accounts/:id/groups/:id
- POST /accounts/:id/groups/:id (invite users to group)
- POST/GET /accounts/:id/collections (create collection)
- PUT/DELETE /accounts/:id/collections/:id (rename/remove collection)
- GET /accounts/:id/collections/:id (collection details, including location)
- GET /accounts/:id/collections/:id/rows (get sub levels of collection)
- GET /accounts/:id/groups/:id (group details)
- POST /accounts/:id/collections/:id/import (import bulk into collection)
- POST /accounts/:id/import (import bulk into account (unsorted into collections))
- GET /accounts/:id (account details)
- POST /accounts/:id/organize (organize based on existing collections)
- GET /accounts/:id/collections/:id/objects (all collection child objects)
- GET /accounts/:id/collections/:id/rows/:id/objects (all collection child objects in row)
- PUT /accounts/:id/objects/:id (modify object properties)
- DELETE /accounts/:id/objects/:id (Remove object)
- GET /health

Mostly the existing nishiki backend API, but more complete, with knowledge of rows and organization)
