// Libraries
import {Dispatch} from 'redux'

// APIs
import {
  getOrganizations as getOrganizationsAPI,
  createOrg as createOrgAPI,
} from 'src/organizations/apis'

// Types
import {AppState, Organization} from 'src/types/v2'

type GetStateFunc = () => Promise<AppState>

export enum ActionTypes {
  SetOrganizations = 'SET_ORGANIZATIONS',
  AddOrg = 'ADD_ORG',
}

export interface SetOrganizations {
  type: ActionTypes.SetOrganizations
  payload: {
    organizations: Organization[]
  }
}

export type Actions = SetOrganizations | AddOrg

export const setOrganizations = (
  organizations: Organization[]
): SetOrganizations => {
  return {
    type: ActionTypes.SetOrganizations,
    payload: {organizations},
  }
}

export interface AddOrg {
  type: ActionTypes.AddOrg
  payload: {
    org: Organization
  }
}

export const addOrg = (org: Organization): AddOrg => ({
  type: ActionTypes.AddOrg,
  payload: {org},
})

// Async Actions

export const getOrganizations = () => async (
  dispatch: Dispatch<SetOrganizations>,
  getState: GetStateFunc
): Promise<void> => {
  try {
    const {
      links: {orgs},
    } = await getState()
    const organizations = await getOrganizationsAPI(orgs)
    dispatch(setOrganizations(organizations))
  } catch (e) {
    console.error(e)
  }
}

export const createOrg = (link: string, org: Partial<Organization>) => async (
  dispatch: Dispatch<AddOrg>
): Promise<void> => {
  try {
    const createdOrg = await createOrgAPI(link, org)
    dispatch(addOrg(createdOrg))
  } catch (e) {
    console.error(e)
  }
}
