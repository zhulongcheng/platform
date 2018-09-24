import React, {PureComponent} from 'react'
import _ from 'lodash'

import InitStep from 'src/onboarding/components/InitStep'
import AdminStep from 'src/onboarding/components/AdminStep'
import CompletionStep from 'src/onboarding/components/CompletionStep'
import OtherStep from 'src/onboarding/components/OtherStep'
import WizardFullScreen from 'src/reusable_ui/components/wizard/WizardFullScreen'
import {ErrorHandling} from 'src/shared/decorators/errors'
import ProgressBar from 'src/reusable_ui/components/wizard/ProgressBar'
import {StepStatus} from 'src/reusable_ui/constants/wizard'

interface Props {
  startStep?: number
  stepStatuses?: StepStatus[]
}

interface State {
  currentStepIndex: number
  stepStatuses: StepStatus[]
}

@ErrorHandling
class OnboardingWizard extends PureComponent<Props, State> {
  public static defaultProps: Partial<Props> = {
    startStep: 0,
    stepStatuses: [
      StepStatus.Incomplete,
      StepStatus.Incomplete,
      StepStatus.Incomplete,
      StepStatus.Incomplete,
    ],
  }

  public stepTitles = ['init', 'admin', 'other', 'complete']
  public steps = [InitStep, AdminStep, OtherStep, CompletionStep]

  constructor(props) {
    super(props)
    this.state = {
      currentStepIndex: props.startStep,
      stepStatuses: props.stepStatuses,
    }
  }

  public render() {
    const {stepStatuses} = this.props
    const {currentStepIndex} = this.state
    return (
      <WizardFullScreen>
        <div className="wizard-container">
          <ProgressBar
            currentStepIndex={currentStepIndex}
            handleSetCurrentStep={this.onSetCurrentStep}
            stepStatuses={stepStatuses}
            stepTitles={this.stepTitles}
          />
        </div>
        <div className="wizard-step--container">
          <div className="wizard-step--child">{this.currentStep}</div>
        </div>
      </WizardFullScreen>
    )
  }

  private get currentStep() {
    const {currentStepIndex} = this.state
    const {stepStatuses} = this.props

    return React.createElement(this.steps[currentStepIndex], {
      stepStatuses,
      stepTitles: this.stepTitles,
      currentStepIndex,
      handleSetCurrentStep: this.onSetCurrentStep,
      handleSetStepStatus: this.onSetStepStatus,
    })
  }

  private onSetCurrentStep = stepNumber => {
    this.setState({currentStepIndex: stepNumber})
  }

  private onSetStepStatus = (index: number, status: StepStatus) => {
    const {stepStatuses} = this.state
    stepStatuses[index] = status
    this.setState({stepStatuses})
  }
}

export default OnboardingWizard
