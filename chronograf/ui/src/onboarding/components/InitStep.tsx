// Libraries
import React, {PureComponent} from 'react'
// Components
import {ErrorHandling} from 'src/shared/decorators/errors'

// Types
import {StepStatus} from 'src/reusable_ui/constants/wizard'

interface Props {
  currentStepIndex: number
  handleSetCurrentStep: (stepNumber: number) => void
  handleSetStepStatus: (index: number, status: StepStatus) => void
  stepStatuses: StepStatus[]
  stepTitles: string[]
}

@ErrorHandling
class InitStep extends PureComponent<Props> {
  public render() {
    return (
      <>
        <div className="auth-logo" />
        <h3 className="wizard-step-title">Welcome to InfluxData </h3>
        <p>"Start using the InfluxData platform in a few easy steps"</p>
        <p>This is Init Step </p>
        <div className="wizard-button-bar">
          <button className="btn btn-md btn-primary" onClick={this.handleNext}>
            Get Started
          </button>
        </div>
      </>
    )
  }

  private handleNext = () => {
    const {handleSetStepStatus, currentStepIndex} = this.props
    handleSetStepStatus(currentStepIndex, StepStatus.Complete)
    this.handleIncrement()
  }

  private handleIncrement = () => {
    const {handleSetCurrentStep, currentStepIndex} = this.props
    handleSetCurrentStep(currentStepIndex + 1)
  }
}

export default InitStep
