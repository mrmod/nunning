import { Grid, Paper, Typography } from "@mui/material";

const PlannedChange = ({count, label, color}) => (<>
    <Typography marginRight={1} color={color || "text.primary"} component="span" variant="h5">{count}</Typography>
    <Typography component="span" variant="h6">{label}</Typography>
</>)

const PlannedChanges = ({updates, deletes, creates, onClick}) => (
    <Grid item container md={12} spacing={2}>
        <Grid item onClick={() => onClick("deletes")}>
            <PlannedChange label={"Deletes"} count={deletes.length} color="warning.main" />
        </Grid>        
        <Grid item onClick={() => onClick("updates")} >
            <PlannedChange label={"Updates"} count={updates.length}  />
        </Grid>
        <Grid item onClick={() => onClick("creates")} >
            <PlannedChange label={"Creates"} count={creates.length}  /> 
        </Grid>
    </Grid>)

export default PlannedChanges