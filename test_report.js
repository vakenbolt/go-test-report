/**
 * @typedef TestStatus
 * @property {string} TestName
 * @property {string} Package
 * @property {number} ElapsedTime
 * @property {Array.<string>} Output
 * @property {boolean} Passed
 */
class TestStatus {}

/**
 * @typedef TestGroupData
 * @type {object}
 * @property {string} FailureIndicator
 * @property {Array.<TestStatus>}
 */
class TestGroupData {}

/**
 * @typedef TestResults
 * @type {Array.<TestGroupData>}
 */
class TestResults extends Array {}

/**
 * @typedef SelectedItems
 * @property {HTMLElement|EventTarget} testResults
 * @property {String} selectedTestGroupColor
 */
class SelectedItems {}

/**
 * @typedef GoTestReportElements
 * @property {TestResults} data
 * @property {HTMLElement} testResultsElem
 * @property {HTMLElement} testGroupListElem
 */
class GoTestReportElements {}


/**
 * Main entry point for GoTestReport.
 * @param {GoTestReportElements} elements
 * @returns {{testResultsClickHandler: testResultsClickHandler}}
 * @constructor
 */
window.GoTestReport = function (elements) {
  const /**@type {SelectedItems}*/ selectedItems = {
    testResults: null,
    selectedTestGroupColor: null
  }

  function addEventData(event) {
    if (event.data == null) {
      event.data = {target: event.target}
    }
    return event
  }


  const goTestReport = {
    /**
     * Invoked when a user clicks on one of the test group div elements.
     * @param {HTMLElement} target The element associated with the test group.
     * @param {boolean} shiftKey If pressed, all of test detail associated to the test group is shown.
     * @param {TestResults} data
     * @param {SelectedItems} selectedItems
     * @param {function(target: Element, data: TestResults)} testGroupListHandler
     */
    testResultsClickHandler: function (target,
                                       shiftKey,
                                       data,
                                       selectedItems,
                                       testGroupListHandler) {
      if (selectedItems.testResults != null) {
        let testResultsElement = /**@type {HTMLElement}*/ selectedItems.testResults
        testResultsElement.style.backgroundColor = selectedItems.selectedTestGroupColor
      }
      const testGroupId = /**@type {number}*/ target.id
      const testResults = /**@type {TestResults}*/ data[testGroupId]['TestResults']
      let testGroupList = /**@type {string}*/ ''
      selectedItems.selectedTestGroupColor = getComputedStyle(target).getPropertyValue('background-color')
      selectedItems.testResults = target
      target.style.backgroundColor = 'black'
      for (let i = 0; i < testResults.length; i++) {
        const testResult = /**@type {TestGroupData}*/ testResults[i]
        const testPassed = /**@type {boolean}*/ testResult.Passed
        const testPassedStatus = /**@type {string}*/ (testPassed) ? '' : 'failed'
        const testId = /**@type {string}*/ target.attributes['id'].value
        testGroupList += `<div class="testGroupRow ${testPassedStatus}" data-groupid="${testId}" data-index="${i}">
        <span class="testStatus ${testPassedStatus}">${(testPassed) ? '&check' : '&cross'};</span>
        <span class="testTitle">${testResult.TestName}</span>
        <span class="testDuration"><span>${testResult.ElapsedTime}s </span>⏱</span>
      </div>`
      }
      const testGroupListElem = elements.testGroupListElem
      testGroupListElem.innerHTML = ''
      testGroupListElem.innerHTML = testGroupList

      if (shiftKey) {
        testGroupListElem.querySelectorAll('.testGroupRow')
                         .forEach((elem) => testGroupListHandler(elem, data))
      } else if (testResults.length === 1) {
        testGroupListHandler(testGroupListElem.querySelector('.testGroupRow'), data)
      }
    },

    /**
     *
     * @param {Element} target
     * @param {TestResults} data
     */
    testGroupListHandler: function (target, data) {
      const attribs = target['attributes']
      if (attribs.hasOwnProperty('data-groupid')) {
        const groupId = /**@type {number}*/ attribs['data-groupid'].value
        const testIndex = /**@type {number}*/ attribs['data-index'].value
        const testStatus = /**@type {TestStatus}*/ data[groupId]['TestResults'][testIndex]
        const testOutputDiv = /**@type {HTMLDivElement}*/ target.querySelector('div.testOutput')

        if (testOutputDiv == null) {
          const testOutputDiv = document.createElement('div')
          testOutputDiv.classList.add('testOutput')
          const consoleSpan = document.createElement('span')
          consoleSpan.classList.add('console')
          const testDetailDiv = document.createElement('div')
          testDetailDiv.classList.add('testDetail')
          const packageNameDiv = document.createElement('div')
          packageNameDiv.classList.add('package')
          packageNameDiv.innerHTML = `<strong>Package:</strong> ${testStatus.Package}`
          const testFileNameDiv = document.createElement('div')
          testFileNameDiv.classList.add('filename')
          testFileNameDiv.innerHTML = `<strong>Filename:</strong> `

          testDetailDiv.insertAdjacentElement('beforeend', packageNameDiv)
          testDetailDiv.insertAdjacentElement('beforeend', testFileNameDiv)
          testOutputDiv.insertAdjacentElement('afterbegin', consoleSpan)
          testOutputDiv.insertAdjacentElement('beforeend', testDetailDiv)
          target.insertAdjacentElement('beforeend', testOutputDiv)

          if (testStatus.Passed) {
            consoleSpan.classList.remove('failed')
          } else {
            consoleSpan.classList.add('failed')
          }
          consoleSpan.textContent = testStatus.Output.join('')
        } else {
          testOutputDiv.remove()
        }
      }
    }
  }

  //+------------------------+
  //|    setup DOM events    |
  //+------------------------+
  elements.testResultsElem
          .addEventListener('click', event =>
            goTestReport.testResultsClickHandler(/**@type {HTMLElement}*/ addEventData(event).data.target,
                                                 event.shiftKey,
                                                 elements.data,
                                                 selectedItems,
                                                 goTestReport.testGroupListHandler))

  elements.testGroupListElem
          .addEventListener('click', event =>
            goTestReport.testGroupListHandler(/**@type {Element}*/ event.target,
                                              elements.data))

  return goTestReport
}