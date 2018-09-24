import {StepStatus} from 'src/reusable_ui/constants/wizard'

export interface Step {
  title: string
  stepStatus: StepStatus
}

export interface NextReturn {
  error: boolean
  payload: any
}
