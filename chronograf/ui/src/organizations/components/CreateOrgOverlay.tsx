import React, {PureComponent, ChangeEvent} from 'react'
import _ from 'lodash'
import {
  Form,
  OverlayBody,
  OverlayHeading,
  OverlayContainer,
  Input,
  Button,
  ComponentColor,
  ComponentStatus,
} from 'src/clockface'

interface Props {
  link: string
  onCloseModal: () => void
}

interface State {
  name: string
  nameInputStatus: ComponentStatus
  errorMessage: string
}

export default class CreateOrgOverlay extends PureComponent<Props, State> {
  constructor(props) {
    super(props)
    this.state = {
      name: '',
      nameInputStatus: ComponentStatus.Default,
      errorMessage: '',
    }
  }

  public render() {
    const {onCloseModal} = this.props
    const {name, nameInputStatus, errorMessage} = this.state

    return (
      <OverlayContainer>
        <OverlayHeading
          title="Create Organization"
          onDismiss={this.props.onCloseModal}
        />
        <OverlayBody>
          <Form>
            <Form.Element label="Name" errorMessage={errorMessage}>
              <Input
                placeholder="Give your organization a name"
                value={name}
                onChange={this.handleChangeName}
                status={nameInputStatus}
              />
            </Form.Element>
            <Form.Footer>
              <Button
                text="Cancel"
                color={ComponentColor.Danger}
                onClick={onCloseModal}
              />
              <Button text="Create" color={ComponentColor.Primary} />
            </Form.Footer>
          </Form>
        </OverlayBody>
      </OverlayContainer>
    )
  }

  private handleChangeName = (e: ChangeEvent<HTMLInputElement>) => {
    const name = e.target.value

    if (_.isEmpty(name)) {
      return this.setState({
        name,
        nameInputStatus: ComponentStatus.Error,
        errorMessage: 'Organization names cannot be empty',
      })
    }

    this.setState({
      name,
      nameInputStatus: ComponentStatus.Valid,
      errorMessage: '',
    })
  }
}
