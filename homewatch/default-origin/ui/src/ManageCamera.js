import { Box, Button } from "@chakra-ui/react"
import { useEffect, useState } from "react"
import { CredentialsPolicy, IsLoaded, IsLoading, IsLoadingError } from "./api"
import Error, { ApiError } from "./Error"

const CamerasStateUrl = "/api/cameras"

const putCameraState = (camera, state) => fetch(CamerasStateUrl, {
    credentials: CredentialsPolicy,
    headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
    },
    method: "PUT",
    body: JSON.stringify({
        camera,
        state,
    })
})
const ManageCamera = ({camera}) => {
    const [cameraState, setCameraState] = useState(null)
    const [loadingState, setLoadingState] = useState(IsLoading)
    const [togglingState, setTogglingState] = useState(IsLoaded)
    const [error, setError] = useState(null)

    const cameraStateUrl = `${CamerasStateUrl}?camera=${camera}`
    
    useEffect(() => {
        fetch(cameraStateUrl, {credentials: CredentialsPolicy})
        .then(response => {
            if (response.ok) {
                return response.json()
            }
            throw(new ApiError(response))
        })
        .then(json => {
            setCameraState(json.cameraState)
            setLoadingState(IsLoaded)
        })
        .catch(err => {
            if (err.status === 403 || err.status === 401) {
                window.location.assign("/")
                return
            }
            setError(err.error())
            setLoadingState(IsLoadingError)
        })
    }, [camera, cameraStateUrl])

    const toggleCameraState = (currentState) => {
        setTogglingState(IsLoading)
        let nextState = "enabled"
        if (currentState.State === "enabled") {
            nextState = "disabled"
        }
        
        putCameraState(currentState.Name, nextState)
        .then(response => {
            if (response.ok) {
                return response.json()
            }
            throw(new ApiError(response))
        })
        .then(json => {
            console.log(`New camera state: ${json}`, json)
            setCameraState(json.cameraState)
            setTogglingState(IsLoaded)
        })
        .catch(err => {
            console.log(`Error setting camera state: ${err}`, err)
            setError(err.error())
            setTogglingState(IsLoadingError)
        })
    }

    switch (loadingState) {
        case IsLoading:
            return <Button
                size={"sm"}
                w={200}
                disabled={true}>
                    Loading {camera} state
                </Button>
        case IsLoadingError:
            return <Error error={error} />
        default:
    }

    return <Box>
        {togglingState === IsLoadingError ? (<Error error={error} />) : null}
        <Button
            size={"sm"}
            w={200}
            onClick={() => toggleCameraState(cameraState)}
            disabled={togglingState === IsLoading}
            colorScheme={cameraState.State === "disabled" ? "gray" : "red"}>
            {cameraState.State === "disabled" ? `Enable ${cameraState.Name}` : `Disable ${cameraState.Name}`}
        </Button>
    </Box>
}

export default ManageCamera