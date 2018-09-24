// Libraries
import React, {PureComponent} from 'react'
import _ from 'lodash'

// Components
import {ErrorHandling} from 'src/shared/decorators/errors'

// constants
import {StepStatus, ConnectorState} from 'src/reusable_ui/constants/wizard'

interface Props {
  currentStepIndex: number
  handleSetCurrentStep: (stepNumber: number) => void
  stepStatuses: StepStatus[]
  stepTitles: string[]
}

@ErrorHandling
class ProgressBar extends PureComponent<Props, null> {
  public render() {
    return (
      <div className="wizard-controller">
        <div className="progress-header">
          <div className="wizard-progress-bar">{this.WizardProgress}</div>
        </div>
      </div>
    )
  }

  private handleSetCurrentStep = i => () => {
    const {handleSetCurrentStep} = this.props
    handleSetCurrentStep(i)
  }

  private get WizardProgress(): JSX.Element {
    const {stepStatuses, stepTitles, currentStepIndex} = this.props

    const lastIndex = stepStatuses.length - 1
    const lastEltIndex = stepStatuses.length - 2

    const progressBar = stepStatuses.reduce((acc, stepStatus, i) => {
      if (i === 0 || i === lastIndex) {
        return [...acc]
      }

      let currentStep = ''
      // STEP STATUS
      if (i === currentStepIndex && stepStatus !== StepStatus.Error) {
        currentStep = 'circle-thick current'
      }

      const stepElt = (
        <div
          key={`stepEle${i}`}
          className="wizard-progress-button"
          onClick={this.handleSetCurrentStep(i)}
        >
          <div className="wizard-progress-title">{stepTitles[i]}</div>
          <span className={`icon ${currentStep || stepStatus}`} />
        </div>
      )

      if (i === lastEltIndex) {
        return [...acc, stepElt]
      }

      // PROGRESS BAR CONNECTOR
      let connectorStatus = ConnectorState.None

      if (i === currentStepIndex && stepStatus !== StepStatus.Error) {
        connectorStatus = ConnectorState.Some
      }
      if (i === lastEltIndex || stepStatus === StepStatus.Complete) {
        connectorStatus = ConnectorState.Full
      }
      const connectorElt = (
        <span
          key={i}
          className={`wizard-progress-connector wizard-progress-connector--${connectorStatus ||
            ConnectorState.None}`}
        />
      )
      return [...acc, stepElt, connectorElt]
    }, [])
    return <>{progressBar}</>
  }
}

export default ProgressBar
