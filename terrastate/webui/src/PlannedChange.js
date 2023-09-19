import { Accordion, AccordionSummary, Grid, Paper, Typography } from "@mui/material";
import { useEffect, useState } from "react";
import ChangeSet from "./ChangeSet";
import { getChangeSet } from "./client";
import CreateDetailCard from "./CreateDetailCard";
import DeleteDetailCard from "./DeleteDetailCard";
import DetailContainer from "./DetailContainer";
import PlannedChanges from "./PlannedChanges";
import UpdateDetailCard from "./UpdateDetailCard";

const showChanges = (changeType) => {
  // TODO
}

const PlannedChange = ({changeSetId}) => {
  const [changeSet, setChangeSet] = useState(null)

  useEffect(() => {
    if (changeSetId) {
      getChangeSet(changeSetId).then(data => {
        if (!data.deletes) {
          data.deletes = []
        }

        if (!data.updates) {
          data.updates = []
        }

        if (!data.creates) {
          data.creates = []
        }
        setChangeSet(data)
      })
    }
  }, [changeSetId])

  if (!changeSet) {
    return <Typography variant="h5">No change selected</Typography>
  }
  
  return <Grid container spacing={2}>
    <Grid item md={12}>
      <Paper sx={{padding: 1, minHeight: 48}}>
        <Grid item md={3}>
          <ChangeSet changeSetId={changeSet.changeSet} />
        </Grid>
        <Grid item md={4}>
          <PlannedChanges
            onClick={showChanges}
            updates={changeSet.updates}
            deletes={changeSet.deletes}
            creates={changeSet.creates} />
        </Grid>
      </Paper>
    </Grid>
    <Grid item md={12}>
      <Accordion sx={{padding: 2, background: "#EEEEEE"}} elevation={0}>
        <AccordionSummary aria-controls="deletes-detail-container" id="deletes-detail-accordian" expandIcon={"+"}>
          <Typography variant={"body1"} color="warning.main">Deleted Resources</Typography>
        </AccordionSummary>
        <DetailContainer>
          {/* Flex: 1 to stretch components to fill the whole width if they're alone on a row */}
          {changeSet.deletes.map((_delete, key) => (<Grid item  flex={1} key={`update-${key}-${_delete.address}`}>
            <DeleteDetailCard {..._delete} />
          </Grid>))}
        </DetailContainer>
      </Accordion>
    </Grid>
    <Grid item md={12}>
      <Accordion sx={{padding: 2, background: "#EEEEEE"}} elevation={0}>
        <AccordionSummary aria-controls="creates-detail-container" id="creates-detail-accordian" expandIcon={"+"}>
          <Typography variant={"body1"}>Created Resources</Typography>
        </AccordionSummary>
        <DetailContainer>
          {changeSet.creates.map((create, key) => (<Grid item key={`create-${key}-${create.address}`}>
            <CreateDetailCard {...create} />
          </Grid>))}
        </DetailContainer>
      </Accordion>
    </Grid>
    <Grid item md={12}>
      <Accordion sx={{padding: 2, background: "#EEEEEE"}} elevation={0}>
        <AccordionSummary aria-controls="updates-detail-container" id="updates-detail-accordian" expandIcon={"+"}>
        <Typography variant={"body1"}>Updated Resources</Typography>
        </AccordionSummary>
        <DetailContainer>
          {changeSet.updates.map((update, key) => (<Grid item key={`update-${key}-${update.address}`}>
            <UpdateDetailCard {...update} changeDiffs={update.change_diffs} />
          </Grid>))}
        </DetailContainer>
      </Accordion>
    </Grid>

  </Grid>
}
export default PlannedChange