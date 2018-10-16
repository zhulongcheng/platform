// Libraries
import React, {Component} from 'react'

// Components
import {Button, ComponentColor, ComponentSize} from 'src/clockface'

// Types
import {Organization} from 'src/types/v2'

interface Props {
  org: Organization
  onDeleteOrg: (org: Organization) => void
}

class DeleteOrgButton extends Component<Props> {
  public render() {
    return (
      <Button
        size={ComponentSize.ExtraSmall}
        color={ComponentColor.Danger}
        text="Delete"
        onClick={this.handleDeleteOrg}
      />
    )
  }

  private handleDeleteOrg = (): void => {
    const {onDeleteOrg, org} = this.props

    onDeleteOrg(org)
  }
}

export default DeleteOrgButton
