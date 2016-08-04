package core

type cmd struct {
	name   string
	params []string
}

type cmdHandler struct {
	name      string
	handler   func(...string) *cmdResult
	argsCount int
}

var commandMap = map[string]cmdHandler{
	//hashes
	"hdel":         {"hdel", doHDel, -2},
	"hexists":      {"hexists", doHExists, 2},
	"hget":         {"hget", doHGet, 2},
	"hgetall":      {"hgetall", doHGetAll, 1},
	"hincrby":      {"hincrby", doHIncrBy, 3},
	"hincrbyfloat": {"hincrbyfloat", doHIncrByFloat, 3},
	"hkeys":        {"hkeys", doHKeys, 1},
	"hlen":         {"hlen", doHLen, 1},
	"hmget":        {"hmget", doHMGet, -2},
	"hmset":        {"hmset", doHMSet, -3},
	"hset":         {"hset", doHSet, 3},
	"hsetnx":       {"hsetnx", doHSetNx, 3},
	//"hscan":        {"hscan", doHScan, -2},
	"hstrlen": {"hstrlen", doHStrlen, 2},
	"hvals":   {"hvals", doHVals, 1},

	//lists
	//"blpop":      {"hvals", doHVals, 1},
	//"brpop":      {"hvals", doHVals, 1},
	//"brpoplpush": {"hvals", doHVals, 1},
	"lindex":    {"lindex", doLIndex, 2},
	"linsert":   {"linsert", doLInsert, 4},
	"llen":      {"llen", doLLen, 1},
	"lpop":      {"lpop", doLPop, 1},
	"lpush":     {"lpush", doLPush, -2},
	"lpushx":    {"lpushx", doLPushX, 2},
	"lrange":    {"lrange", doLRange, 3},
	"lrem":      {"lrem", doLRem, 3},
	"lset":      {"lset", doLSet, 3},
	"ltrim":     {"lTrim", doLTrim, 3},
	"rpop":      {"rpop", doRPop, 1},
	"rpoplpush": {"rpoplpush", doRPopLPush, 2},
	"rpush":     {"rpush", doRPush, -2},
	"rpushx":    {"rpushx", doRPushX, 2},

	//server
	"flushall": {"flushall", doFlushAll, 0},
	"flushdb":  {"flushdb", doFlushDb, 0},

	//sets
	"sadd":        {"sadd", doSAdd, -2},
	"scard":       {"scard", doSCard, 1},
	"sdiff":       {"sdiff", doSDiff, -1},
	"sdiffstore":  {"sdiffstore", doSDiffStore, -2},
	"sinter":      {"sinter", doSInter, -1},
	"sinterstore": {"sinterstore", doSInterStore, -2},
	"sismember":   {"sismember", doSIsMember, 2},
	"smembers":    {"smembers", doSMembers, 1},
	"smove":       {"smove", doSMove, 3},
	//"spop":        {"spop", doSPop, -1},
	//"srandmember": {"srandmember", doSRandMember, -1},
	"srem":        {"srem", doSRem, -2},
	"sunion":      {"sunion", doSUnion, -1},
	"sunionstore": {"sunionstore", doSUnionStore, -2},
	//"sscan": {"sscan", doSScan, -2},

	//sorted sets
	"zadd":             {"zadd", doZAdd, -3},
	"zcard":            {"zcard", doZCard, 1},
	"zcount":           {"zcount", doZCount, 3},
	"zincrby":          {"zincrby", doZIncrBy, 3},
	"zinterstore":      {"zinterstore", doZInterStore, -3},
	"zlexcount":        {"zlexcount", doZLexCount, 3},
	"zrange":           {"zrange", doZRange, -3},
	"zrangebylex":      {"zrangebylex", doZRangeByLex, -3},
	"zrevrangebylex":   {"zrevrangebylex", doZRevRangeByLex, -3},
	"zrangebyscore":    {"zrangebyscore", doZRangeByScore, -3},
	"zrank":            {"zrank", doZRank, 2},
	"zrem":             {"zrem", doZRem, -2},
	"zremrangebylex":   {"zremrangebylex", doZRemRrangeByLex, 3},
	"zremrangebyrank":  {"zremrangebyrank", doZRemRangeByRank, 3},
	"zremrangebyscore": {"zremrangebyscore", doZRemRangeByScore, 3},
	"zrevrange":        {"zrevrange", doZRevRange, -3},
	"zrevrangebyscore": {"zrevrangebyscore", doZRevRangeByScore, -3},
	"zrevrank":         {"zrevrank", doZRevRank, 2},
	"zsocre":           {"zscore", doZScore, 2},
	"zunionstore":      {"zunionstore", doZUnionStore, -3},
	//"zscan":            {"zscan", doZScan, -2},

	//strings
	"append":   {"append", doAppend, 2},
	"bitcount": {"bitcount", doBitCount, -1},
	//"bitfield": {"bitfiled", doBitField, -1},
	"bitop":       {"bitop", doBitOp, -3},
	"bitpos":      {"bitpos", doBitPos, -2},
	"decr":        {"decr", doDecr, 1},
	"decrby":      {"decrby", doDecrBy, 2},
	"get":         {"get", doGet, 1},
	"getbit":      {"getbit", doGetBit, 2},
	"getrange":    {"getrange", doGetRange, 3},
	"getset":      {"getset", doGetSet, 2},
	"incr":        {"incr", doIncr, 1},
	"incrby":      {"incrby", doIncrBy, 2},
	"incrbyfloat": {"incrbyfloat", doIncrByFloat, 2},
	"mget":        {"mget", doMGet, -1},
	"mset":        {"mset", doMSet, -2},
	"msetnx":      {"msetnx", doMSetNx, -2},
	"psetex":      {"psetex", doPSetEx, 3},
	"set":         {"set", doSet, -2},
	"setbit":      {"setbit", doSetBit, 3},
	"setex":       {"setex", doSetEx, 3},
	"setnx":       {"setnx", doSetNx, 2},
	"setrange":    {"setrange", doSetRange, 3},
	"strlen":      {"strlen", doStrlen, 1},
}
