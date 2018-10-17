import {Organization} from 'src/types/v2'
import {ActionTypes, Actions} from 'src/organizations/actions'

const defaultState = []

export default (state = defaultState, action: Actions): Organization[] => {
  switch (action.type) {
    case ActionTypes.SetOrganizations:
      return [...action.payload.organizations]
    case ActionTypes.AddOrg:
      return [...state, {...action.payload.org}]
    default:
      return state
  }
}
