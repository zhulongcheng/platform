// Libraries
import React, {SFC} from 'react'

interface Props {
  children: any
}

const WizardFullScreen: SFC<Props> = (props: Props) => {
  return (
    <div className="wizard--full-screen">
      {props.children}
      <p className="auth-credits">
        Made by <span className="icon cubo-uniform" />InfluxData
      </p>
      <div className="auth-image" />
    </div>
  )
}

export default WizardFullScreen
