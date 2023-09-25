import { Spinner, Stack } from "@chakra-ui/react"
import { useEffect, useState } from "react"
import { IsLoaded, IsLoading, IsLoadingError, CredentialsPolicy, } from "./api"
import Error from "./Error"
import RawDatapointsData from "./RawDatapointsData"
import {rmUrlEnvironment} from "./util"

const Video = ({camera, transcodeDataUrl}) => {
    const [videoUrl, setVideoUrl] = useState(null)
    const [transcodeData, setTranscodeData] = useState({})
    const [loadingState, setLoadingState] = useState(IsLoading)
    const [error, setError] = useState(null)
    
    useEffect(() => {
        fetch(transcodeDataUrl, {credentials: CredentialsPolicy})
        .then(response => {
            if (response.ok) {
                return response.json()
            }
            response.text().then(text => {throw(text)})
        })
        .then(json => {
            setTranscodeData(json)
            
            if (json.urls.length > 1) {
                console.log(`Found more than one video url for ${transcodeDataUrl}`)
            }
            // Trim the leading environment specifier
            setVideoUrl(rmUrlEnvironment(json.urls[0]))
            setLoadingState(IsLoaded)
        })
        .catch(err => {
            setError(err)
            setLoadingState(IsLoadingError)
        })
    }, [transcodeDataUrl])
    if (loadingState === IsLoading) {
        return <Spinner></Spinner>
    }
    if (loadingState === IsLoadingError) {
        return <Error error={error} />
    }

    return <Stack>
        <RawDatapointsData datapoints={transcodeData} />
        <video controls={true} width={896} height={414}>
            <source
                src={videoUrl}
                autoPlay={false}
                type="video/mp4"
            />
        </video>

    </Stack>
}

export default Video