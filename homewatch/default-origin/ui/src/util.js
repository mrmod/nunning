const rmUrlEnvironment = (url) => url.split("/").slice(1).join("/")

// TODO: Utilize Day.js
const asDate = (s) => {
    const [_, year, month, day, hour, minute, second] = s.match(/^(\d{4})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})/)
    return {
        year,
        month,
        day,
        hour: parseInt(hour) <= 12 ? hour : parseInt(hour)%12,
        minute,
        second,
        ampm: parseInt(hour) < 12 ? "am" : "pm"
    }
}

const onlyImage = (v) => v.split("/keyframes/")[1].split("_")[1]
// Sorts image 10.jpeg before image 11.jpeg
const byKeyframeUrl = (a, b) => onlyImage(a).localeCompare(onlyImage(b))

const DatapointDateFormat = "YYYYMMDDHHmmss"
const datapointComparator = (a, b) => b.DateTime.localeCompare(a.DateTime)


const getHour = (dateTime) => parseInt(dateTime.slice(8,10))
const getMinute = (dateTime) => parseInt(dateTime.slice(10, 12))
const getDateTuple = (dateTime) => dateTime.slice(0,8)

// toUnix: Converts a  Datapoint{DateTime: $field} to unix seconds since epoch
const toUnix = (dp) => {
    const dateTime = dp.DateTime
    const d = new Date(
        parseInt(dateTime.slice(0,4)),    // year
        parseInt(dateTime.slice(4, 6))-1, // month
        parseInt(dateTime.slice(6, 8)),   // dom
        getHour(dateTime),   // hour
        getMinute(dateTime), // minute
        parseInt(dateTime.slice(12, 14)), //Second
    )
    return d.getTime()
}

// byDateTime : Ascending (a - b), Descending (b-a) by Date
const byDateTime = (a, b) => toUnix(b) - toUnix(a)

const getMinuteBin = (datapoint) => {
    const minute = getMinute(datapoint.DateTime)
    if (minute < 15) return 0
    if (minute < 30) return 1
    if (minute < 45) return 2
    return 3
}

/* invertedBin: Bins Datapoints into
    bins = [
        // YYYYmmDD Bin [
            // 0/15/30/45m Bin [
                // 0-24h Bin [ Datapoint ]
            ]
        ]
    ]
    Time: O(n)
    Space: O(n)
*/
const invertedBin = (datapoints) => {
    datapoints.sort(byDateTime)

    const dateIndex = {}
    const dateSet = new Set()

    // Fill in date bins with datapoints
    for (let dp of datapoints) {
        let hour = getHour(dp.DateTime)
        let date = getDateTuple(dp.DateTime)
        let minBin = getMinuteBin(dp)

        // Index by date
        if (!dateIndex[date]) {
        // The minBin array ends up being a reference to the same
        // array using the below fill style
        // dateIndex[date] = new Array(4).fill(new Array(24).fill([]))

        // So we use this style so
        // minBin is a unique array
        dateIndex[date] = new Array(4)
        for (let m = 0 ; m < 4; m++ ) {
            dateIndex[date][m] = new Array(24)
            for (let mm = 0 ; mm < 24; mm++) {
            dateIndex[date][m][mm] = []
            }
        }
        }

        dateIndex[date][minBin][hour] = dateIndex[date][minBin][hour].concat([dp])
        dateSet.add(date)
    }
    return {bins: dateIndex, dateSet: Array.from(dateSet)}
}

const yearMonthDay = (dateTuple) => {
    const year = parseInt(dateTuple.slice(0, 4))
    const month = parseInt(dateTuple.slice(4, 6)) -1
    const day = parseInt(dateTuple.slice(6, 8))
    return new Date(year, month, day).toLocaleDateString()
}

export {yearMonthDay, invertedBin, toUnix, asDate, rmUrlEnvironment, byKeyframeUrl,DatapointDateFormat, datapointComparator}