const { expectEvent, constants, time, expectRevert } = require('@openzeppelin/test-helpers');

const Course = artifacts.require('CourseMock');

contract('Course', accounts => {
    const [teacher, evaluator, student, other] = accounts;
    let course = null;
    const digest1 = web3.utils.keccak256(web3.utils.toHex('cert1'));
    const digest2 = web3.utils.keccak256(web3.utils.toHex('cert2'));
    const digest3 = web3.utils.keccak256(web3.utils.toHex('cert3'));
    const courseDigest = web3.utils.keccak256(web3.utils.toHex('course1'));

    describe('constructor', () => {
        it('should successfully deploy the contract', async () => {
            let beginTimestamp = (await time.latest()).add(time.duration.seconds(1));
            let endTimestamp = beginTimestamp.add(await time.duration.hours(1));
            course = await Course.new([teacher, evaluator], 2, beginTimestamp.toString(), endTimestamp.toString());
            (await course.isOwner(teacher)).should.equal(true);
            (await course.isOwner(evaluator)).should.equal(true);
            assert(course.quorum(), 2);
        });
    });

    describe('CRUD operations', () => {

        beforeEach(async () => {
            let beginTimestamp = (await time.latest()).add(time.duration.seconds(1));
            let endTimestamp = beginTimestamp.add(await time.duration.hours(1));
            course = await Course.new([teacher, evaluator], 2, beginTimestamp, endTimestamp);
            await time.increase(time.duration.seconds(1));
        });

        it('should add a new student', async () => {
            const { logs } = await course.addStudent(student, { from: teacher });

            expectEvent.inLogs(logs, 'StudentAdded', {
                student: student,
                requester: teacher
            });

            (await course.enrolledStudents(student)).should.equal(true);
        });

        it('should not add a student twice', async () => {
            await course.addStudent(student, { from: teacher });
            await expectRevert(
                course.addStudent(student, { from: evaluator }),
                "Course: student already registered in this course"
            );
        });

        it('should verify if a student is enrolled in the course', async () => {
            (await course.isEnrolled(student)).should.equal(false);
            await course.addStudent(student, { from: teacher });
            (await course.isEnrolled(student)).should.equal(true);
        });

        it('should not allow zero address', async () => {
            await expectRevert(
                course.addStudent(constants.ZERO_ADDRESS, { from: teacher }),
                "Course: student is the zero address"
            );
        });

        it('should allow an owner to remove an enrolled student', async () => {
            await course.addStudent(student, { from: evaluator });
            (await course.enrolledStudents(student)).should.equal(true);

            const { logs } = await course.removeStudent(student, { from: teacher });

            expectEvent.inLogs(logs, 'StudentRemoved', {
                student: student,
                requester: teacher
            });

            (await course.enrolledStudents(student)).should.equal(false);
        });

        it('should revert if try to remove an unregistered student', async () => {
            await expectRevert(
                course.removeStudent(student, { from: teacher }),
                "Course: student not registered"
            );
        });

        it('should allow a student to renounce the course', async () => {
            await course.addStudent(student, { from: teacher });
            (await course.enrolledStudents(student)).should.equal(true);

            const { logs } = await course.renounceCourse({ from: student });

            expectEvent.inLogs(logs, 'StudentRemoved', {
                student: student,
                requester: student
            });

            (await course.enrolledStudents(student)).should.equal(false);
        });
    });

    describe('Certification', () => {

        beforeEach(async () => {
            let beginTimestamp = (await time.latest()).add(time.duration.seconds(1));
            let endTimestamp = beginTimestamp.add(await time.duration.hours(1));
            course = await Course.new([teacher, evaluator], 2, beginTimestamp, endTimestamp);
            await time.increase(time.duration.seconds(1));
        });

        it('should issue a credential for a enrolled student', async () => {
            await course.addStudent(student, { from: teacher });
            await course.issue(student, digest1, { from: teacher });
        });

        it('should not issue a credential for a non-enrolled address', async () => {
            await expectRevert(
                course.issue(other, digest1, { from: teacher }),
                'Course: student not registered'
            );
        });

        it('should create a course certificate based on all valid exams of a subject', async () => {
            await course.enrollStudents([student, other]);

            for (d of [digest1, digest2, digest3]) {
                await course.issue(student, d, { from: teacher });
                await course.issue(student, d, { from: evaluator });
                await course.confirmProof(d, { from: student });
                await time.increase(time.duration.seconds(1));

                (await course.certified(d)).should.equal(true);
            }

            await course.issue(student, courseDigest, { from: teacher });
            await course.issue(student, courseDigest, { from: evaluator });
            await course.confirmProof(courseDigest, { from: student });

            await time.increase(time.duration.hours(1));
            (await course.hasEnded()).should.equal(true);

            const aggregated = await course.aggregate.call(student);
            let expected = web3.utils.keccak256(web3.eth.abi.encodeParameter('bytes32[]', [digest1, digest2, digest3, courseDigest]));
            (aggregated).should.equal(expected);
        });
    });
});
