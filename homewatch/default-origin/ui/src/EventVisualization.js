import { useEffect, useState } from "react"
import { VegaLite } from "react-vega"

const spec = {
    // $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    description: "Events per time window",
    data: {name: "values"},
    view: { fill: "aliceblue"},
    mark: {
        type: "area",
        stroke: "red",
        fill: "tomato",
    },
    format: {
        parse: {"date": "%Y-%m-%d %H:%M %Z"}
    },
    encoding: {
        y: {
            start: 0,
            // title: "Events",
            title: false,
            field: "count",
            type: "quantitative"
        },
        x: {
            // title: "EventTime",
            title: false,
            field: "date",
            sort: "descending",
            type: "temporal",
            timeUnit: "hoursminutes",
            axis: {format: "%I:%M %p"},
        },
    }
}
const transformBinnedData = (binnedData) => {
    let values = []
    binnedData.binIds.forEach((binDate, binId) => {
        if (binId !== undefined) {
            values.push({
                count: binnedData.bins[binId].length,
                date: binDate.format("YYYY-M-D HH:mmZ"),
            })
        }
    })
    
    return {values}
}

const EventVisualization = ({binnedData}) => {
    const [data, setData] = useState({values: []})
    useEffect(() => {
        if (binnedData.bins.length > 0 ) {
            setData(transformBinnedData(binnedData))
        }
    }, [binnedData])
    
    return <VegaLite
        height={80}
        width={860}
        data={data}
        spec={spec}
    />
}

export default EventVisualization