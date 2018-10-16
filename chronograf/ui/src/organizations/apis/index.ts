import AJAX from 'src/utils/ajax'

import {Member, Bucket, Dashboard, Task} from 'src/types/v2'

export const getMembers = async (url: string): Promise<Member[]> => {
  try {
    const {data} = await AJAX({
      url,
    })

    return data.members
  } catch (error) {
    console.error('Could not get members for org', error)
    throw error
  }
}

export const getBuckets = async (url: string): Promise<Bucket[]> => {
  try {
    const {data} = await AJAX({
      url,
    })

    return data.buckets
  } catch (error) {
    console.error('Could not get buckets for org', error)
    throw error
  }
}

export const getDashboards = async (url: string): Promise<Dashboard[]> => {
  try {
    const {data} = await AJAX({
      url,
    })

    return data.dashboards
  } catch (error) {
    console.error('Could not get buckets for org', error)
    throw error
  }
}

export const getTasks = async (url: string): Promise<Task[]> => {
  try {
    const {data} = await AJAX({
      url,
    })

    return data.tasks
  } catch (error) {
    console.error('Could not get tasks for org', error)
    throw error
  }
}
