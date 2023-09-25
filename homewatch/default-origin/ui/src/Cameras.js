import { Tabs, TabList, TabPanels, Tab, TabPanel } from '@chakra-ui/react'
import CameraContainer from './CameraContainer'

const deriveCameras = () => {
    if (process.env.REACT_APP_CAMERAS === undefined) {
        return ["DemoCamera1", "DemoCamera2"]
    }
    return process.env.REACT_APP_CAMERAS.split(",").map(c => c.trim())
}

const cameras = deriveCameras()
/*
Cameras

Container of the main user interface for the application.

It's a horizontally tabbed-interface, one per camera

*/
const Cameras = () => (<Tabs isLazy>
    <TabList>
        {cameras.map((camera) => (<Tab key={`tab-${camera}`}>
            <h1>{camera}</h1>
        </Tab>))}
    </TabList>
    <TabPanels>
        {cameras.map((camera) => (<TabPanel p={0} ml={3} key={`panel-${camera}`} >
            <CameraContainer camera={camera} />
        </TabPanel>))}
    </TabPanels>
</Tabs>)


export default Cameras