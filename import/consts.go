package main

const (
	BtcNetwork  = 1
	EthNetwork  = 0
	importQuery = "MERGE (t:Transaction {hash:$hash}) CREATE (t2:Transaction) SET t.hash = $hash, t.timestamp = datetime($timestamp), " +
		"t.totalBTC = toFloat($total_amount), t.totalUSD = toFloat($total_usd),t.flowBTC = toFloat($flow_btc), " +
		"t.flowUSD = toFloat($flow_usd) MERGE (a:Address {id:$from_address}) " +
		"CREATE (a)-[:SENT {valueBTC:toFloat($from_value), valueUSD: toFloat($from_value_usd)}]->(t) " +
		"MERGE (b:Address {id:$to_address}) CREATE (t)-[:SENT {valueBTC: toFloat($to_value), valueUSD: toFloat($to_value_usd)}]->(b)"
)
