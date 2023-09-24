import { Text, Box, Accordion, Spinner, Grid, GridItem, Flex, Button} from "@chakra-ui/react"
import { useEffect, useState } from "react"
import Error, {ApiError} from "./Error"

import {CredentialsPolicy, IsLoaded, IsLoading, IsLoadingError} from "./api"
import { invertedBin, toUnix, yearMonthDay } from "./util"
const eventStyle = (h) => ({
    width: 24,
    height: 24,
    marginRight: 0.2,
    textAlign: "center",
    fontFamily: "roboto sans-serif",
    background: h > 0 ? 'red' : h < 0 ? '#eeeeee' : 'white'
})

const hours = () => new Array(24).fill(true)

const HourLabels = () => (<Box elevation={0} style={{display: "flex"}}>
    {hours().map((h, key) => (<Box
       
        elevation={0}
        key={`a-${key}`}
        style={eventStyle(-1)}>
        {key < 11 ? `${key+1}a` : key == 11 ? `12p`: `${key-11}p`}
    </Box>))}
    <Box elevation={0} style={eventStyle(-1)}></Box>
</Box>)

// binData: Matrix sized 4x24 where the first range is the 0,15,30,45 range
//          and the second rank is the hour where the length is the count of 
//          items for that MinuteBin[Hour] coordinate
const DataRow = ({binData, minBin, onClick}) => (<Box elevation={0} style={{display: "flex"}}>
    {binData.map((h, key) => (<Box
       
        elevation={1}
        onClick={() => onClick(minBin, key, h)}
        key={`a-${key}`}
        style={eventStyle(h.length)}>
            {h.length}
    </Box>))}
    <Box elevation={0} style={eventStyle(-1)}>{minBin*15}</Box>
    </Box>)

const apiUrl = process.env.REACT_APP_API_URL
const rmUrlEnvironment = (url) => url.split("/").slice(1).join("/")
const videoUrl = (davKey) => {
    const video = rmUrlEnvironment(davKey).slice(0, rmUrlEnvironment(davKey).indexOf("[M]"))+".mp4"

    return `${apiUrl}/${video}`
}
const CameraBox = ({camera}) => {
    const [displayData, setDisplayData] = useState([])
    const DatapointsUrl = `/api/datapoints?camera=${camera}`
    const [pages, setPages] = useState(1)
    const [page, setPage] = useState(1)
    const [datapoints, setDatapoints] = useState([])
    const [loadingState, setLoadingState] = useState(IsLoading)
    const [error, setError] = useState(null)
    useEffect(() => {
        let offset = `&pages=${pages}`
        
        setLoadingState(IsLoading)
        fetch(DatapointsUrl+offset, {credentials: CredentialsPolicy})
        .then(response => {
            if (response.ok) {
                return response.json()
            }
            throw(new ApiError(response))
        })
        .then(json => {
            setLoadingState(IsLoaded)
            // setDatapoints(json.datapoints.sort(datapointComparator))
            setDatapoints(invertedBin(json.datapoints))
        })
        .catch(apiError => {

            if (apiError.status === 403 || apiError.status === 401) {
                // window.location.assign("/")
                console.log("403: ", apiError)
                return
            }
            setLoadingState(IsLoadingError)
            if (apiError.error) {
                setError(apiError.error())
            } else {
                console.log(apiError)
                return
            }
        })
    }, [camera, DatapointsUrl, pages])

    const goBack = () => {
        setPages(pages+1)
        setPage(page+1)
    } 
    const goForward = () => {
        setPages(pages > 2 ? pages-1 : 1)
    }
    const displayBinData = (minBin, hourBin, binData) => setDisplayData(binData)
    switch (loadingState) {
        case IsLoading:
            return <Box>
                {/* <ManageCamera camera={camera} /> */}
                <Spinner></Spinner>
            </Box>
        case IsLoadingError:
            return <Box>
                {/* <ManageCamera camera={camera} /> */}
                <Error error={error} />
            </Box>
        default:
    }
    return <Grid>
        <GridItem >
        <Button onClick={goBack}>Older by 48 hours</Button>
        <Button disabled={pages === 1} onClick={goForward}>Newer by 48 hours</Button>
        </GridItem>
        {datapoints.dateSet.map((binId, key) => (<GridItem key={`bin-${binId}`}>
            <Text fontSize={"sm"}>{yearMonthDay(binId)}</Text>
            <HourLabels/>
            {datapoints.bins[binId].map((hourBin, hourKey) => (
                <DataRow
                    key={`hbin-${hourKey}`}
                    minBin={hourKey}
                    binData={hourBin}
                    onClick={displayBinData}
                />
            ))}
        </GridItem>))}
        <GridItem w={"100%"}>
            <Accordion >
                <Flex flexWrap={"wrap"} w={"100%"}>
                    {displayData.map((datapoint, key) => (<Box w={540} key={`dpk-${datapoint.DateTime}`}>
                        <Box style={{padding: 1}}>
                        <Text fontSize="lg">{new Date(toUnix(datapoint)).toLocaleString()}</Text>
                        <video controls>
                            <source
                            src={videoUrl(datapoint.DavKey)}
                            autoPlay={false}
                            type="video/mp4" />
                        </video>
                        </Box>
                    </Box>))}
                </Flex>
            </Accordion>
        </GridItem>
    </Grid>
}
export default CameraBox