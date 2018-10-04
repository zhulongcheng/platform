import AJAX from 'src/utils/ajax'

import {Organization} from 'src/types'

export const getOrgs = async (
  url: string
): Promise<{orgs: Organization[]; links: {}}> => {
  try {
    const {data} = await AJAX({
      url,
    })

    return data
  } catch (error) {
    throw error
  }
}
