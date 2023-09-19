const plansUrl = "/plans/"
const diffsUrl = "/diffs/"

const getPlans = () => fetch(plansUrl, {
  headers: {
    "Content-type": "application/json",
  }
}).then(response => (response.ok ? response.json() : {"error": `request failed ${response.status}`}))

const getChangeSet = (changeSetId) => fetch(diffsUrl+changeSetId, {
  headers: {
    "Content-type": "application/json",
  }
}).then(response => (response.ok ? response.json() : {"error": `request failed ${response.status}`}))

export {getPlans, getChangeSet}