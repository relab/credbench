const { expectEvent, BN, time, expectRevert, ether } = require('@openzeppelin/test-helpers');

const Faculty = artifacts.require('FacultyMock');
const Course = artifacts.require('CourseMock');

contract('Faculty', accounts => {
    const [dean, adm, teacher, evaluator, student, other] = accounts;
    let faculty, courseStarts, courseEnds = null;
    const semester = web3.utils.keccak256(web3.utils.toHex('spring2020'));

    describe('constructor', () => {
        it('should successfully deploy the contract', async () => {
            faculty = await Faculty.new([dean, adm], 2);
            (await faculty.isOwner(dean)).should.equal(true);
            (await faculty.isOwner(adm)).should.equal(true);
            assert(faculty.quorum(), 2);
        });
    });

    describe('course creation', () => {
        beforeEach(async () => {
            faculty = await Faculty.new([dean, adm], 2);
            faculty.setBalance({ value: ether('1') });
            courseStarts = (await time.latest()).add(time.duration.seconds(1));
            courseEnds = courseStarts.add(await time.duration.hours(1));
        });

        it('should not create a course from a unauthorized address', async () => {
            await expectRevert(
                faculty.createCourse(semester, [teacher, evaluator], 2, courseStarts, courseEnds, { from: other }),
                'Owners: sender is not an owner'
            );
        });

        it('should retrieve a course by semester', async () => {
            let course = await Course.new([teacher, evaluator], 2, courseStarts, courseEnds, { from: adm });
            await faculty.addCourse(course.address, semester);

            (await faculty.coursesBySemester(semester, 0)).should.equal(course.address);
        });

        it('should create a course', async () => {
            const { logs } = await faculty.createCourse(semester, [teacher, evaluator], 2, courseStarts, courseEnds, { from: adm });

            expectEvent.inLogs(logs, 'CourseCreated', {
                createdBy: adm,
                semester: semester,
                courseAddress: await faculty.coursesBySemester(semester, 0),
                quorum: new BN(2)
            });
        });

        it('course should be an issuer instance', async () => {
            await faculty.createCourse(semester, [teacher, evaluator], 2, courseStarts, courseEnds, { from: adm });

            let courseAddress = await faculty.coursesBySemester(semester, 0);
            (await faculty.isIssuer(courseAddress)).should.equal(true);
        });
    });

    describe('issuing diploma', () => {
        beforeEach(async () => {
            faculty = await Faculty.new([dean, adm], 2);
            faculty.setBalance({ value: ether('1') });
        });

        it('should issue a diploma', async () => {
            let coursesAddress = [];
            var certs = [];
            let numberOfCourses = 2;
            let numberOfExams = 2;
            for (i = 0; i < numberOfCourses; i++) {
                // Create the course
                courseStarts = (await time.latest()).add(time.duration.seconds(1 + i));
                courseEnds = courseStarts.add(await time.duration.hours(1));
                await faculty.createCourse(semester, [teacher, evaluator], 2, courseStarts, courseEnds, { from: adm });
                // Start the course
                await time.increase(time.duration.seconds(1 + i));
                // Get the course instance
                let course = await Course.at(await faculty.coursesBySemester(semester, i));
                coursesAddress.push(course.address);

                // Add student
                await course.addStudent(student, { from: teacher });
                // Add exam certificates
                for (j = 0; j < numberOfExams; j++) {
                    let examDigest = web3.utils.keccak256(web3.utils.toHex(`course${i}-exam${j}`));
                    // issue exams certificate
                    await course.registerCredential(student, examDigest, { from: teacher });
                    await course.registerCredential(student, examDigest, { from: evaluator });
                    await course.confirmCredential(examDigest, { from: student });
                    await time.increase(time.duration.seconds(1));
                }

                // Add course certificate
                let courseDigest = web3.utils.keccak256(web3.utils.toHex(`course${i}`));
                await course.registerCredential(student, courseDigest, { from: teacher });
                await course.registerCredential(student, courseDigest, { from: evaluator });
                await course.confirmCredential(courseDigest, { from: student });
                await time.increase(time.duration.seconds(1));

                certs.push(await course.digestsBySubject(student));
            }
            await time.increase(time.duration.hours(1));

            for (let cAddress of coursesAddress) {
                let course = await Course.at(cAddress);
                await course.aggregateCredentials(student, { from: teacher });
            }
            // Aggregate courses certs
            var expectedCerts = certs.map(c => web3.utils.keccak256(web3.eth.abi.encodeParameter('bytes32[]', c)));

            // Add diploma digest
            let diplomaDigest = web3.utils.keccak256(web3.utils.toHex('diploma'));
            // Add new diploma to expected certs
            expectedCerts.push(diplomaDigest);

            // Create new root
            let expectedRoot = web3.utils.keccak256(web3.eth.abi.encodeParameter('bytes32[]', expectedCerts));

            // Finish courses
            await time.increase(time.duration.hours(1));

            // issue a diploma
            await faculty.registerCredential(student, diplomaDigest, expectedRoot, coursesAddress);//, { from: adm });
            // FIXME: possible bug when passing from address as parameter

            let diploma = (await faculty.digestsBySubject(student))[0];
            (diploma).should.equal(diplomaDigest);
        });
    });

    describe('verifying diploma', () => { });
});