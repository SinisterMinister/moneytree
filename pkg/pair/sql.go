package pair

const orderStatsQuery string = `
with pairs as (
	select
		*,
		("btcReturn" * "firstPrice") + "usdReturn" - ("buyFees" + "sellFees" + "revFees") as "totalReturn"
	from (
		select
			*,
			case 
				when "status" = 'SUCCESS' then "buyFilled" - "sellFilled"
				when "status" = 'REVERSED' and "direction" = 'DOWN' then "revFilled" - "buyFilled" - "sellFilled"
				when "status" = 'REVERSED' and "direction" = 'UP' then "buyFilled" - "sellFilled" - "revFilled"
				else 0
			end as "btcReturn",
			case 
				when "status" = 'SUCCESS' then "sellPrice" * "sellFilled" - "buyPrice" * "buyFilled"
				when "status" = 'REVERSED' and "direction" = 'DOWN' then ("sellPrice" * "sellFilled" - "buyPrice" * "buyFilled") - "revPrice" * "revFilled"
				when "status" = 'REVERSED' and "direction" = 'UP' then ("revPrice" * "revFilled") - ABS("sellPrice" * "sellFilled" - "buyPrice" * "buyFilled")
				else 0
			end as "usdReturn"
		from (
			select
				*,
				tsrange("createdAt", "endedAt") as "timeslot",
				case when "direction" = 'UP' then "firstStatus" else "secondStatus" end as "buyStatus",
				case when "direction" = 'UP' then "firstPrice" else "secondPrice" end as "buyPrice",
				case when "direction" = 'UP' then "firstQty" else "secondQty" end as "buyQty",
				case when "direction" = 'UP' then "firstFilled" else "secondFilled" end as "buyFilled",
				case when "direction" = 'UP' then "firstFees" else "secondFees" end as "buyFees",
				
				case when "direction" = 'DOWN' then "firstStatus" else "secondStatus" end as "sellStatus",
				case when "direction" = 'DOWN' then "firstPrice" else "secondPrice" end as "sellPrice",
				case when "direction" = 'DOWN' then "firstQty" else "secondQty" end as "sellQty",
				case when "direction" = 'DOWN' then "firstFilled" else "secondFilled" end as "sellFilled",
				case when "direction" = 'DOWN' then "firstFees" else "secondFees" end as "sellFees"
			from (
				select
					uuid,
					(data->>'createdAt')::timestamp as "createdAt",
					case
						when data->>'endedAt' = '0001-01-01T00:00:00Z' then LOCALTIMESTAMP
						when data->>'status' = 'OPEN' then LOCALTIMESTAMP
						else (data->>'endedAt')::timestamp
					end as "endedAt",
					data->>'direction' as "direction",
					data->>'status' as "status",
					data->'firstOrder'->>'status' as "firstStatus",
					data->'secondOrder'->>'status' as "secondStatus",
					data->'reversalOrder'->>'status' as "revStatus",
					(data->'firstRequest'->>'price')::decimal as "firstPrice",
					(data->'secondRequest'->>'price')::decimal as "secondPrice",
					(data->'reversalOrder'->'request'->>'price')::decimal as "revPrice",
					(data->'firstRequest'->>'quantity')::decimal as "firstQty",
					(data->'secondRequest'->>'quantity')::decimal as "secondQty",
					(data->'reversalRequest'->>'funds')::decimal as "revFunds",
					(data->'firstOrder'->>'filled')::decimal as "firstFilled",
					(data->'secondOrder'->>'filled')::decimal as "secondFilled",
					(data->'reversalOrder'->>'filled')::decimal as "revFilled",
					(data->'reversalOrder'->>'fees')::decimal as "revFees",
					(data->'secondOrder'->>'fees')::decimal as "secondFees",
					(data->'firstOrder'->>'fees')::decimal as "firstFees"
				from
					orderpairs
			) as raw_pairs
		) as pair_returns
		where
			timeslot && tsrange(?, ?)
	) as total_returns
)
`
