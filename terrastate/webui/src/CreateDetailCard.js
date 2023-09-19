const { Card, CardContent, Typography, Accordion, AccordionSummary, AccordionDetails } = require("@mui/material");

const stripType = (type, address) => address.replace(`${type}.`, "")

const CreateDetailCard = ({address, values, type, name}) => (<Card sx={{minWidth: 320}}>
    <CardContent>
        <Typography sx={{fontSize: 14}} color="text.secondary" gutterBottom>
            {type}
        </Typography>
        <Typography variant="h5" component="div">
            {values.name ? values.name : stripType(type, address)}
        </Typography>
        <Typography sx={{ mb: 1.5}} color="text.secondary">
            {address}
        </Typography>
        <Accordion>
            <AccordionSummary expandIcon={"+"} >
                <Typography variant="body1">Resource Definition</Typography>
            </AccordionSummary>
            <AccordionDetails>
                <pre>
                    {JSON.stringify(values, null, "  ")}
                </pre>
            </AccordionDetails>
        </Accordion>
    </CardContent>
</Card>)

export default CreateDetailCard