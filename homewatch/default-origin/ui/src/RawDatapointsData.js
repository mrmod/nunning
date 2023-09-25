import {Accordion, AccordionButton, AccordionIcon, AccordionItem, AccordionPanel, Box, Code} from "@chakra-ui/react"


const RawDatapointsData = ({datapoints}) => {
    if (process.env.REACT_APP_ENVIRONMENT !== undefined) {
        return null
    }
    return <Accordion allowToggle>
        <AccordionItem>
            <h3>
                <AccordionButton>
                    <Box flex={"1"} textAlign={"left"}>
                        Raw data
                    </Box>
                    <AccordionIcon />
                </AccordionButton>
            </h3>
            <AccordionPanel pb={4}>
                <Code>
                    {JSON.stringify(datapoints, null, "  ")}
                </Code>
            </AccordionPanel>
        </AccordionItem>
    </Accordion>
}

export default RawDatapointsData