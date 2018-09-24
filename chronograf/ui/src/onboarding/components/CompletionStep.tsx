// Libraries
import React, {PureComponent} from 'react'
import {withRouter, WithRouterProps} from 'react-router'

// Components
import {ErrorHandling} from 'src/shared/decorators/errors'

// Types
import {StepStatus} from 'src/reusable_ui/constants/wizard'

interface Props extends WithRouterProps {
  currentStepIndex: number
  handleSetCurrentStep: (stepNumber: number) => void
  handleSetStepStatus: (index: number, status: StepStatus) => void
  stepStatuses: StepStatus[]
  stepTitles: string[]
}

@ErrorHandling
class CompletionStep extends PureComponent<Props> {
  public render() {
    return (
      <>
        <div className="auth-logo" />
        <h3 className="wizard-step-title">Setup Complete! </h3>
        <p>"Start using the InfluxData platform in a few easy steps"</p>
        <p>This is Init Step </p>
        <div className="wizard-button-bar">
          <button
            className="btn btn-md btn-default"
            onClick={this.handleDecrement}
          >
            Back
          </button>
          <button
            className="btn btn-md btn-primary"
            onClick={this.handleComplete}
          >
            Go to Status Dashboard
          </button>
        </div>
      </>
    )
  }

  private handleDecrement = () => {
    const {handleSetCurrentStep, currentStepIndex} = this.props
    handleSetCurrentStep(currentStepIndex - 1)
  }

  private handleComplete = () => {
    const {router} = this.props
    router.push(`/manage-sources`)
  }
}

export default withRouter(CompletionStep)
