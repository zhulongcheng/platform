// Libraries
import React, {PureComponent} from 'react'
import {Link} from 'react-router'

const settingsLinks = [
  {link: 'user/settings', title: 'Settings'},
  {link: 'user/tokens', title: 'Tokens'},
]

export default class SupportLinks extends PureComponent {
  public render() {
    return (
      <ul className="link-list">
        {settingsLinks.map(({link, title}) => (
          <li key={title}>
            <Link to={link}>{title}</Link>
          </li>
        ))}
      </ul>
    )
  }
}
