package usecases

import (
	"time"

	"github.com/nishiki/backend/domain/entities"
)

// --- TypedValue helpers ---

// TV creates a text TypedValue.
func TV(s string) entities.TypedValue {
	return entities.TypedValue{Type: entities.PropertyTypeText, Val: s}
}

// Props is shorthand for building a TypedValue properties map from key-value pairs.
// Values can be entities.TypedValue or plain strings (auto-wrapped as text).
func Props(kvs ...any) map[string]entities.TypedValue {
	m := make(map[string]entities.TypedValue, len(kvs)/2)
	for i := 0; i < len(kvs)-1; i += 2 {
		key := kvs[i].(string)
		switch v := kvs[i+1].(type) {
		case entities.TypedValue:
			m[key] = v
		case string:
			m[key] = TV(v)
		default:
			m[key] = entities.TypedValue{Val: v}
		}
	}
	return m
}

// --- Entity builders with functional options ---

// TestObject builds a minimal reconstructed Object. Override fields via opts.
func NewTestObject(opts ...func(*objectOpts)) *entities.Object {
	o := objectOpts{
		name:  "Test Object",
		props: map[string]entities.TypedValue{},
		tags:  []string{},
	}
	for _, fn := range opts {
		fn(&o)
	}
	objName, _ := entities.NewObjectName(o.name)
	return entities.ReconstructObject(
		o.id.orNew(), objName, entities.NewObjectDescription(o.desc),
		entities.ObjectTypeGeneral, "", o.quantity, o.unit,
		o.props, o.tags, "", o.expiresAt,
		time.Now(), time.Now(),
	)
}

type objectOpts struct {
	id        optionalID[entities.ObjectID]
	name      string
	desc      string
	unit      string
	quantity  *float64
	props     map[string]entities.TypedValue
	tags      []string
	expiresAt *time.Time
}

func ObjName(n string) func(*objectOpts)           { return func(o *objectOpts) { o.name = n } }
func ObjDesc(d string) func(*objectOpts)           { return func(o *objectOpts) { o.desc = d } }
func ObjID(id entities.ObjectID) func(*objectOpts) { return func(o *objectOpts) { o.id.set(id) } }
func ObjProps(p map[string]entities.TypedValue) func(*objectOpts) {
	return func(o *objectOpts) { o.props = p }
}
func ObjTags(t ...string) func(*objectOpts)      { return func(o *objectOpts) { o.tags = t } }
func ObjUnit(u string) func(*objectOpts)         { return func(o *objectOpts) { o.unit = u } }
func ObjQuantity(q float64) func(*objectOpts)    { return func(o *objectOpts) { o.quantity = &q } }
func ObjExpiresAt(t time.Time) func(*objectOpts) { return func(o *objectOpts) { o.expiresAt = &t } }

// TestContainer builds a minimal reconstructed Container. Override fields via opts.
func NewTestContainer(opts ...func(*containerOpts)) *entities.Container {
	o := containerOpts{
		name:    "Test Container",
		ctype:   entities.ContainerTypeGeneral,
		objects: []entities.Object{},
	}
	for _, fn := range opts {
		fn(&o)
	}
	name, _ := entities.NewContainerName(o.name)
	return entities.ReconstructContainer(
		o.id.orNew(), o.collectionID.orNew(), name, o.ctype,
		o.parentID, nil, o.groupID,
		o.objects, o.location,
		nil, nil, nil, nil,
		time.Now(), time.Now(),
	)
}

type containerOpts struct {
	id           optionalID[entities.ContainerID]
	collectionID optionalID[entities.CollectionID]
	name         string
	ctype        entities.ContainerType
	parentID     *entities.ContainerID
	groupID      *entities.GroupID
	objects      []entities.Object
	location     string
}

func CtrName(n string) func(*containerOpts) { return func(o *containerOpts) { o.name = n } }
func CtrID(id entities.ContainerID) func(*containerOpts) {
	return func(o *containerOpts) { o.id.set(id) }
}
func CtrCollectionID(id entities.CollectionID) func(*containerOpts) {
	return func(o *containerOpts) { o.collectionID.set(id) }
}
func CtrGroupID(id *entities.GroupID) func(*containerOpts) {
	return func(o *containerOpts) { o.groupID = id }
}
func CtrObjects(objs ...entities.Object) func(*containerOpts) {
	return func(o *containerOpts) { o.objects = objs }
}
func CtrLocation(l string) func(*containerOpts) { return func(o *containerOpts) { o.location = l } }

// TestCollection builds a minimal reconstructed Collection. Override fields via opts.
func NewTestCollection(opts ...func(*collectionOpts)) *entities.Collection {
	o := collectionOpts{
		name:       "Test Collection",
		objectType: entities.ObjectTypeGeneral,
		containers: []entities.Container{},
		tags:       []string{},
	}
	for _, fn := range opts {
		fn(&o)
	}
	name, _ := entities.NewCollectionName(o.name)
	return entities.ReconstructCollection(
		o.id.orNew(), o.userID.orNew(), o.groupID, name, nil,
		o.objectType, o.containers, o.tags, o.location, o.schema,
		time.Now(), time.Now(),
	)
}

type collectionOpts struct {
	id         optionalID[entities.CollectionID]
	userID     optionalID[entities.UserID]
	groupID    *entities.GroupID
	name       string
	objectType entities.ObjectType
	containers []entities.Container
	tags       []string
	location   string
	schema     *entities.PropertySchema
}

func ColName(n string) func(*collectionOpts) { return func(o *collectionOpts) { o.name = n } }
func ColID(id entities.CollectionID) func(*collectionOpts) {
	return func(o *collectionOpts) { o.id.set(id) }
}
func ColUserID(id entities.UserID) func(*collectionOpts) {
	return func(o *collectionOpts) { o.userID.set(id) }
}
func ColGroupID(id *entities.GroupID) func(*collectionOpts) {
	return func(o *collectionOpts) { o.groupID = id }
}
func ColContainers(c ...entities.Container) func(*collectionOpts) {
	return func(o *collectionOpts) { o.containers = c }
}
func ColTags(t ...string) func(*collectionOpts)  { return func(o *collectionOpts) { o.tags = t } }
func ColLocation(l string) func(*collectionOpts) { return func(o *collectionOpts) { o.location = l } }
func ColSchema(s *entities.PropertySchema) func(*collectionOpts) {
	return func(o *collectionOpts) { o.schema = s }
}

// TestGroup builds a minimal reconstructed Group. Override fields via opts.
func NewTestGroup(opts ...func(*groupOpts)) *entities.Group {
	o := groupOpts{
		name: "Test Group",
		desc: "Test group description",
	}
	for _, fn := range opts {
		fn(&o)
	}
	name, _ := entities.NewGroupName(o.name)
	return entities.ReconstructGroup(
		o.id.orNew(), name, entities.NewGroupDescription(o.desc),
		time.Now(), time.Now(),
	)
}

type groupOpts struct {
	id   optionalID[entities.GroupID]
	name string
	desc string
}

func GrpName(n string) func(*groupOpts)          { return func(o *groupOpts) { o.name = n } }
func GrpID(id entities.GroupID) func(*groupOpts) { return func(o *groupOpts) { o.id.set(id) } }
func GrpDesc(d string) func(*groupOpts)          { return func(o *groupOpts) { o.desc = d } }

// --- generic optional ID ---

type entityID interface {
	entities.ObjectID | entities.ContainerID | entities.CollectionID | entities.UserID | entities.GroupID
}

type optionalID[T entityID] struct {
	val *T
}

func (o *optionalID[T]) set(id T) { o.val = &id }

// orNew returns the set value, or generates a new one. This uses a type switch
// on a pointer to the zero value to dispatch to the correct New* function.
func (o optionalID[T]) orNew() T {
	if o.val != nil {
		return *o.val
	}
	var zero T
	switch any(&zero).(type) {
	case *entities.ObjectID:
		id := entities.NewObjectID()
		return any(id).(T)
	case *entities.ContainerID:
		id := entities.NewContainerID()
		return any(id).(T)
	case *entities.CollectionID:
		id := entities.NewCollectionID()
		return any(id).(T)
	case *entities.UserID:
		id := entities.NewUserID()
		return any(id).(T)
	case *entities.GroupID:
		id := entities.NewGroupID()
		return any(id).(T)
	default:
		panic("unsupported ID type")
	}
}
