import { FormControl, Grid, InputLabel, MenuItem, Select } from "@mui/material";
import { useEffect, useState } from "react";

import PlannedChange from "./PlannedChange";

import {getPlans} from "./client"

const byCreatedAt = (a, b) => a.created_at - b.created_at
const App = () => {
  const [plans, setPlans] = useState([])
  const [changeSetId, setChangeSetId] = useState("")

  useEffect(() => {
    getPlans().then(plans => setPlans(plans.sort(byCreatedAt)))
  }, [])

  const selectChangeSetId = (event) => setChangeSetId(event.target.value)
  return <Grid item>
    <Grid item md={12} container>
      <FormControl fullWidth>
        <InputLabel id="select-change-set-id-label">Change Set</InputLabel>
        <Select
          defaultValue={plans[0] ? plans[0].change_set_id : null}
          value={changeSetId}
          label="ChangeSet"
          labelId="select-change-set-id-label"
          id="select-change-set-id"
          onChange={selectChangeSetId}
        >
          {plans.map((plan, key) => (<MenuItem
            key={`plan-${plan.created_at}-${plan.change_set_id}`}
            value={plan.change_set_id}
            >
            {new Date(plan.created_at*1000).toISOString()}
          </MenuItem>))}
        </Select>
      </FormControl>
    </Grid>
    <Grid item md={12}>
      <PlannedChange changeSetId={changeSetId} />
    </Grid>
  </Grid>
}

export default App